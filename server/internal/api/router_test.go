package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/api"
	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/ocr"
)

type noopOCR struct{}

func (noopOCR) Parse(context.Context, []byte) (ocr.ParsedReceipt, error) {
	return ocr.ParsedReceipt{}, nil
}

type noopStorage struct{}

func (noopStorage) Upload(context.Context, string, []byte, string) (string, error) {
	return "https://example.test/x", nil
}

// This exercises the real router wiring (internal/api/router.go) rather
// than testing handler/middleware logic in isolation — it would have
// caught, e.g., a route accidentally registered outside the RequireAuth
// group, or CORS not being mounted globally.
func newTestRouter(t *testing.T) *gin.Engine {
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
		&models.Household{}, &models.User{}, &models.Product{}, &models.Receipt{},
		&models.WarrantyRule{}, &models.WarrantyClaim{}, &models.DeviceToken{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	h := handlers.New(db, config.Config{JWTSecret: "test-secret", UploadsDir: t.TempDir()}, noopOCR{}, noopStorage{})
	return api.NewRouter(h)
}

func TestNewRouter_HealthEndpointIsPublic(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("body = %q, want it to contain \"status\":\"ok\"", rec.Body.String())
	}
}

func TestNewRouter_ProtectedRoutesRejectUnauthenticatedRequests(t *testing.T) {
	tests := []struct {
		method, path string
	}{
		{http.MethodGet, "/households/me"},
		{http.MethodGet, "/products"},
		{http.MethodPost, "/products"},
		{http.MethodGet, "/warranty-rules/resolve"},
		{http.MethodPost, "/devices"},
		{http.MethodGet, "/products/some-id/claims"},
	}

	router := newTestRouter(t)
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestNewRouter_AuthRoutesArePublic(t *testing.T) {
	router := newTestRouter(t)
	// An empty body fails validation (400), not auth (401) — proving this
	// route is reachable without a token at all.
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (validation error, not 401)", rec.Code, http.StatusBadRequest)
	}
}

func TestNewRouter_CORSIsMountedGlobally(t *testing.T) {
	router := newTestRouter(t)
	// Even a protected route must answer an OPTIONS preflight before
	// RequireAuth ever runs.
	req := httptest.NewRequest(http.MethodOptions, "/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS headers on a protected route's preflight response")
	}
}

func TestNewRouter_ValidTokenReachesProtectedHandler(t *testing.T) {
	router := newTestRouter(t)

	regBody := `{"email":"a@example.com","password":"supersecret1","full_name":"Test User"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("registration failed: status %d, body %s", regRec.Code, regRec.Body.String())
	}

	var reg struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(regRec.Body.Bytes(), &reg); err != nil {
		t.Fatalf("failed to decode registration response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/households/me", nil)
	req.Header.Set("Authorization", "Bearer "+reg.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}
