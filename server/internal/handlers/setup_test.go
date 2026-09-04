package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/auth"
	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/middleware"
	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/ocr"
)

// fakeOCR lets each test control exactly what "OCR" returns instead of
// depending on ocr.StubProvider's fixed output.
type fakeOCR struct {
	result ocr.ParsedReceipt
	err    error
}

func (f *fakeOCR) Parse(context.Context, []byte) (ocr.ParsedReceipt, error) {
	return f.result, f.err
}

// fakeStorage avoids touching disk: it just fabricates a URL from the key.
type fakeStorage struct {
	err error
}

func (f *fakeStorage) Upload(_ context.Context, key string, _ []byte, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "https://fake-storage.test/" + key, nil
}

// testSetup wires an authenticated router (products, receipts, claims, and
// households routes behind the real RequireAuth middleware) against an
// isolated in-memory SQLite database, plus one ready-made household/user to
// act as.
type testSetup struct {
	router      *gin.Engine
	db          *gorm.DB
	ocrProvider *fakeOCR
	userID      uuid.UUID
	householdID uuid.UUID
	token       string
	storage     *fakeStorage
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.Household{}, &models.User{}, &models.Product{},
		&models.Receipt{}, &models.WarrantyRule{}, &models.WarrantyClaim{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	household := models.Household{Name: "Test Household", InviteCode: "TESTCODE"}
	if err := db.Create(&household).Error; err != nil {
		t.Fatalf("failed to seed household: %v", err)
	}
	user := models.User{Email: "owner@example.com", PasswordHash: "x", FullName: "Owner", HouseholdID: household.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	fakeOCRProvider := &fakeOCR{}
	fakeStorageProvider := &fakeStorage{}
	h := handlers.New(db, config.Config{JWTSecret: testJWTSecret}, fakeOCRProvider, fakeStorageProvider)

	router := gin.New()
	authed := router.Group("/", middleware.RequireAuth(testJWTSecret))
	authed.POST("/products", h.CreateProduct)
	authed.GET("/products", h.ListProducts)
	authed.GET("/products/:id", h.GetProduct)
	authed.PUT("/products/:id", h.UpdateProduct)
	authed.POST("/receipts", h.UploadReceipt)
	authed.GET("/receipts/:id", h.GetReceipt)
	authed.POST("/products/:id/claims", h.CreateClaim)
	authed.GET("/products/:id/claims", h.ListClaims)
	authed.GET("/households/me", h.GetMyHousehold)

	token, err := auth.GenerateAccessToken(testJWTSecret, user.ID, household.ID)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	return &testSetup{
		router:      router,
		db:          db,
		ocrProvider: fakeOCRProvider,
		userID:      user.ID,
		householdID: household.ID,
		token:       token,
		storage:     fakeStorageProvider,
	}
}

// createOtherHousehold seeds a second, unrelated household+user and returns
// a valid access token for it — used to test cross-household authorization.
func (s *testSetup) createOtherHousehold(t *testing.T) (token string, householdID uuid.UUID) {
	t.Helper()
	household := models.Household{Name: "Other Household", InviteCode: "OTHERCODE"}
	if err := s.db.Create(&household).Error; err != nil {
		t.Fatalf("failed to seed other household: %v", err)
	}
	user := models.User{Email: "outsider@example.com", PasswordHash: "x", FullName: "Outsider", HouseholdID: household.ID}
	if err := s.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed other user: %v", err)
	}
	tok, err := auth.GenerateAccessToken(testJWTSecret, user.ID, household.ID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return tok, household.ID
}

// addHouseholdMember seeds a second user in the caller's own household —
// used to test household member listings.
func (s *testSetup) addHouseholdMember(t *testing.T, email, fullName string) models.User {
	t.Helper()
	user := models.User{Email: email, PasswordHash: "x", FullName: fullName, HouseholdID: s.householdID}
	if err := s.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed household member: %v", err)
	}
	return user
}

func (s *testSetup) seedRule(t *testing.T, category, brand string, months int) {
	t.Helper()
	rule := models.WarrantyRule{Category: category, Brand: brand, DurationMonths: months, Source: "default"}
	if err := s.db.Create(&rule).Error; err != nil {
		t.Fatalf("failed to seed warranty rule: %v", err)
	}
}

func doJSONAs(t *testing.T, router *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = *bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doMultipartAs(t *testing.T, router *gin.Engine, method, path, token, fieldName, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("failed to write form file data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// noBodyMultipart posts an empty multipart form with no file part, to test
// the "missing file" validation path.
func doMultipartNoFileAs(t *testing.T, router *gin.Engine, method, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
