package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/api/idtoken"

	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/models"
)

// fakeGoogleValidate lets each test control exactly what a "Google ID
// token" claims to be, without needing a real token signed by Google.
func fakeGoogleValidate(t *testing.T, claims map[string]any, err error) func() {
	t.Helper()
	original := handlers.ValidateGoogleIDToken
	handlers.ValidateGoogleIDToken = func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		if err != nil {
			return nil, err
		}
		return &idtoken.Payload{Claims: claims}, nil
	}
	return func() { handlers.ValidateGoogleIDToken = original }
}

func TestGoogleLogin_CreatesNewUserAndHousehold(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeGoogleValidate(t, map[string]any{
		"sub": "google-sub-1", "email": "new@example.com", "name": "משתמש חדש",
	}, nil)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "fake"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var user models.User
	if err := s.db.Where("email = ?", "new@example.com").First(&user).Error; err != nil {
		t.Fatalf("expected a new user to be created: %v", err)
	}
	if user.GoogleID != "google-sub-1" {
		t.Errorf("GoogleID = %q, want %q", user.GoogleID, "google-sub-1")
	}
	if user.FullName != "משתמש חדש" {
		t.Errorf("FullName = %q, want %q", user.FullName, "משתמש חדש")
	}

	var householdCount int64
	s.db.Model(&models.Household{}).Where("id = ?", user.HouseholdID).Count(&householdCount)
	if householdCount != 1 {
		t.Errorf("expected a household to be created for the new user, found %d", householdCount)
	}
}

func TestGoogleLogin_LogsInExistingGoogleLinkedUser(t *testing.T) {
	s := newTestSetup(t)
	existing := models.User{Email: "linked@example.com", GoogleID: "google-sub-2", FullName: "כבר קיים", HouseholdID: s.householdID}
	if err := s.db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to seed existing Google-linked user: %v", err)
	}

	restore := fakeGoogleValidate(t, map[string]any{
		"sub": "google-sub-2", "email": "linked@example.com", "name": "כבר קיים",
	}, nil)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "fake"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var count int64
	s.db.Model(&models.User{}).Where("google_id = ?", "google-sub-2").Count(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 user with this google_id, found %d (should not create a duplicate)", count)
	}
}

func TestGoogleLogin_LinksGoogleIDToExistingEmailPasswordAccount(t *testing.T) {
	s := newTestSetup(t)
	existing := models.User{Email: "hasPassword@example.com", PasswordHash: "x", FullName: "יש לי כבר סיסמה", HouseholdID: s.householdID}
	if err := s.db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to seed existing password-based user: %v", err)
	}

	restore := fakeGoogleValidate(t, map[string]any{
		"sub": "google-sub-3", "email": "hasPassword@example.com", "name": "יש לי כבר סיסמה",
	}, nil)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "fake"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var linked models.User
	if err := s.db.Where("email = ?", "hasPassword@example.com").First(&linked).Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if linked.GoogleID != "google-sub-3" {
		t.Errorf("GoogleID = %q, want %q (should link to the existing account)", linked.GoogleID, "google-sub-3")
	}
	if linked.ID != existing.ID {
		t.Error("expected the same user record to be reused, not a new one created")
	}
}

func TestGoogleLogin_InvalidTokenReturns401(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeGoogleValidate(t, nil, context.DeadlineExceeded)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "garbage"})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGoogleLogin_MissingRequiredClaimsReturns401(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeGoogleValidate(t, map[string]any{"sub": "google-sub-4"}, nil) // no email claim
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "fake"})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGoogleLogin_MissingIDTokenReturns400(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGoogleLogin_ReturnsServiceUnavailableWhenNotConfigured(t *testing.T) {
	s := newTestSetup(t)
	original := s.h.Cfg.GoogleOAuthClientID
	s.h.Cfg.GoogleOAuthClientID = ""
	defer func() { s.h.Cfg.GoogleOAuthClientID = original }()

	rec := doJSONAs(t, s.router, http.MethodPost, "/auth/google", "", map[string]any{"id_token": "fake"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
