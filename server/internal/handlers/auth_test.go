package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/auth"
	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/models"
)

const testJWTSecret = "test-secret"

// newTestRouter gives each test an isolated in-memory SQLite database and a
// Gin engine wired with just the auth routes (no rate limiting/CORS — those
// are middleware concerns, not the handlers' behavior).
func newTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
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

	if err := db.AutoMigrate(&models.Household{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	h := handlers.New(db, config.Config{JWTSecret: testJWTSecret}, nil, nil)

	router := gin.New()
	router.POST("/auth/register", h.Register)
	router.POST("/auth/login", h.Login)
	router.POST("/auth/refresh", h.RefreshToken)
	return router, db
}

func doJSON(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
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
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
}

type authResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func registerUser(t *testing.T, router *gin.Engine, email, fullName, inviteCode string) (*httptest.ResponseRecorder, authResponse) {
	t.Helper()
	rec := doJSON(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":       email,
		"password":    "supersecret1",
		"full_name":   fullName,
		"invite_code": inviteCode,
	})
	var resp authResponse
	if rec.Code == http.StatusCreated {
		decodeJSON(t, rec, &resp)
	}
	return rec, resp
}

func TestRegister_CreatesHouseholdAndUser(t *testing.T) {
	router, db := newTestRouter(t)

	rec, resp := registerUser(t, router, "michal@example.com", "מיכל כהן", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected non-empty access and refresh tokens")
	}
	if resp.User.Email != "michal@example.com" {
		t.Errorf("User.Email = %q, want %q", resp.User.Email, "michal@example.com")
	}
	if resp.User.HouseholdID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("expected a real household_id to be assigned")
	}

	var household models.Household
	if err := db.First(&household, "id = ?", resp.User.HouseholdID).Error; err != nil {
		t.Fatalf("expected a household row to exist: %v", err)
	}
	if household.CreatedBy != resp.User.ID {
		t.Errorf("household.CreatedBy = %v, want %v", household.CreatedBy, resp.User.ID)
	}
	if household.InviteCode == "" {
		t.Error("expected a non-empty invite code to be generated")
	}

	var stored models.User
	if err := db.First(&stored, "id = ?", resp.User.ID).Error; err != nil {
		t.Fatalf("expected a user row to exist: %v", err)
	}
	if stored.PasswordHash == "" || stored.PasswordHash == "supersecret1" {
		t.Error("password must be hashed, not stored in plaintext (or left empty)")
	}

	claims, err := auth.ParseAccessToken(testJWTSecret, resp.AccessToken)
	if err != nil {
		t.Fatalf("access token did not parse: %v", err)
	}
	if claims.UserID != resp.User.ID || claims.HouseholdID != resp.User.HouseholdID {
		t.Error("access token claims don't match the returned user")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "supersecret1", "full_name": "Ron"}},
		{"invalid email", map[string]string{"email": "not-an-email", "password": "supersecret1", "full_name": "Ron"}},
		{"missing password", map[string]string{"email": "ron@example.com", "full_name": "Ron"}},
		{"password too short", map[string]string{"email": "ron@example.com", "password": "short", "full_name": "Ron"}},
		{"missing full_name", map[string]string{"email": "ron@example.com", "password": "supersecret1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _ := newTestRouter(t)
			rec := doJSON(t, router, http.MethodPost, "/auth/register", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestRegister_DuplicateEmailIsRejected(t *testing.T) {
	router, db := newTestRouter(t)

	rec, _ := registerUser(t, router, "dup@example.com", "First", "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("first registration failed: status %d, body %s", rec.Code, rec.Body.String())
	}

	rec2, _ := registerUser(t, router, "dup@example.com", "Second", "")
	if rec2.Code == http.StatusCreated {
		t.Fatal("second registration with the same email should not succeed")
	}

	var count int64
	db.Model(&models.User{}).Where("email = ?", "dup@example.com").Count(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 user with that email, found %d", count)
	}
}

func TestRegister_InviteCodeJoinsExistingHousehold(t *testing.T) {
	router, db := newTestRouter(t)

	_, first := registerUser(t, router, "a@example.com", "A", "")
	var household models.Household
	if err := db.First(&household, "id = ?", first.User.HouseholdID).Error; err != nil {
		t.Fatalf("failed to load household: %v", err)
	}

	rec, second := registerUser(t, router, "b@example.com", "B", household.InviteCode)
	if rec.Code != http.StatusCreated {
		t.Fatalf("second registration failed: status %d, body %s", rec.Code, rec.Body.String())
	}
	if second.User.HouseholdID != first.User.HouseholdID {
		t.Errorf("second user's household_id = %v, want %v (should join, not create a new household)",
			second.User.HouseholdID, first.User.HouseholdID)
	}

	var householdCount int64
	db.Model(&models.Household{}).Count(&householdCount)
	if householdCount != 1 {
		t.Errorf("expected exactly 1 household, found %d", householdCount)
	}
}

func TestRegister_InviteCodeRejectsThirdMember(t *testing.T) {
	router, db := newTestRouter(t)

	_, first := registerUser(t, router, "a@example.com", "A", "")
	var household models.Household
	db.First(&household, "id = ?", first.User.HouseholdID)

	rec2, _ := registerUser(t, router, "b@example.com", "B", household.InviteCode)
	if rec2.Code != http.StatusCreated {
		t.Fatalf("second registration should succeed: status %d, body %s", rec2.Code, rec2.Body.String())
	}

	rec3, _ := registerUser(t, router, "c@example.com", "C", household.InviteCode)
	if rec3.Code == http.StatusCreated {
		t.Fatal("a third member should not be allowed to join (2-member cap)")
	}

	var memberCount int64
	db.Model(&models.User{}).Where("household_id = ?", household.ID).Count(&memberCount)
	if memberCount != 2 {
		t.Errorf("expected exactly 2 members in the household, found %d", memberCount)
	}
}

func TestRegister_InvalidInviteCodeIsRejected(t *testing.T) {
	router, db := newTestRouter(t)

	rec, _ := registerUser(t, router, "solo@example.com", "Solo", "NOTAREALCODE")
	if rec.Code == http.StatusCreated {
		t.Fatal("registration with a bogus invite code should not succeed")
	}

	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount != 0 {
		t.Errorf("expected no user to be created, found %d", userCount)
	}
	var householdCount int64
	db.Model(&models.Household{}).Count(&householdCount)
	if householdCount != 0 {
		t.Errorf("expected no household to be created either, found %d", householdCount)
	}
}

func TestLogin_Success(t *testing.T) {
	router, _ := newTestRouter(t)
	registerUser(t, router, "login@example.com", "Login Test", "")

	rec := doJSON(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "login@example.com",
		"password": "supersecret1",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp authResponse
	decodeJSON(t, rec, &resp)
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected non-empty tokens on successful login")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	router, _ := newTestRouter(t)
	registerUser(t, router, "login2@example.com", "Login Test", "")

	rec := doJSON(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "login2@example.com",
		"password": "wrong-password",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLogin_UnknownEmailReturnsGenericError(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doJSON(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "whatever123",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}

	var errResp errorResponse
	decodeJSON(t, rec, &errResp)
	if errResp.Error != "invalid email or password" {
		t.Errorf("error = %q, want the same generic message used for a wrong password (avoid leaking which emails exist)", errResp.Error)
	}
}

func TestRefreshToken_IssuesNewAccessToken(t *testing.T) {
	router, _ := newTestRouter(t)
	_, reg := registerUser(t, router, "refresh@example.com", "Refresh Test", "")

	rec := doJSON(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": reg.RefreshToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp struct {
		AccessToken string `json:"access_token"`
	}
	decodeJSON(t, rec, &resp)
	if resp.AccessToken == "" {
		t.Fatal("expected a non-empty access token")
	}

	claims, err := auth.ParseAccessToken(testJWTSecret, resp.AccessToken)
	if err != nil {
		t.Fatalf("new access token did not parse: %v", err)
	}
	if claims.UserID != reg.User.ID || claims.HouseholdID != reg.User.HouseholdID {
		t.Error("new access token claims don't match the original user")
	}
}

func TestRefreshToken_RejectsAnAccessTokenUsedAsRefresh(t *testing.T) {
	router, _ := newTestRouter(t)
	_, reg := registerUser(t, router, "wrongtype@example.com", "Test", "")

	// Pass the access token where a refresh token belongs.
	rec := doJSON(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": reg.AccessToken,
	})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (an access token must not work as a refresh token)", rec.Code, http.StatusUnauthorized)
	}
}

func TestRefreshToken_RejectsGarbage(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doJSON(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "not-a-real-token",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
