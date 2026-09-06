package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/models"
)

// fakeExchangeGmailCode lets each test control exactly what "exchanging an
// authorization code" returns, without needing a real code issued by
// Google's consent screen.
func fakeExchangeGmailCode(t *testing.T, token *oauth2.Token, gmailAddress string, err error) func() {
	t.Helper()
	original := handlers.ExchangeGmailCode
	handlers.ExchangeGmailCode = func(context.Context, *handlers.Handler, string, string, string) (*oauth2.Token, string, error) {
		if err != nil {
			return nil, "", err
		}
		return token, gmailAddress, nil
	}
	return func() { handlers.ExchangeGmailCode = original }
}

func TestConnectGmail_StoresConnectionOnSuccess(t *testing.T) {
	s := newTestSetup(t)
	expiry := time.Now().Add(time.Hour)
	restore := fakeExchangeGmailCode(t, &oauth2.Token{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		Expiry:       expiry,
	}, "user@gmail.com", nil)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{
		"code":         "auth-code",
		"redirect_uri": "http://localhost:8081",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var conn models.GmailConnection
	if err := s.db.Where("user_id = ?", s.userID).First(&conn).Error; err != nil {
		t.Fatalf("expected a stored GmailConnection: %v", err)
	}
	if conn.GmailAddress != "user@gmail.com" {
		t.Errorf("GmailAddress = %q, want %q", conn.GmailAddress, "user@gmail.com")
	}
	if conn.EncryptedAccessToken == "" || conn.EncryptedAccessToken == "access-1" {
		t.Error("expected the access token to be stored encrypted, not empty or in plaintext")
	}
	if conn.EncryptedRefreshToken == "" || conn.EncryptedRefreshToken == "refresh-1" {
		t.Error("expected the refresh token to be stored encrypted, not empty or in plaintext")
	}
}

func TestConnectGmail_ReconnectingUpdatesExistingRow(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeExchangeGmailCode(t, &oauth2.Token{AccessToken: "a1", RefreshToken: "r1", Expiry: time.Now().Add(time.Hour)}, "first@gmail.com", nil)
	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{"code": "c1", "redirect_uri": "http://localhost:8081"})
	if rec.Code != http.StatusOK {
		t.Fatalf("first connect status = %d, want %d", rec.Code, http.StatusOK)
	}
	restore()

	restore2 := fakeExchangeGmailCode(t, &oauth2.Token{AccessToken: "a2", RefreshToken: "r2", Expiry: time.Now().Add(time.Hour)}, "second@gmail.com", nil)
	defer restore2()
	rec2 := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{"code": "c2", "redirect_uri": "http://localhost:8081"})
	if rec2.Code != http.StatusOK {
		t.Fatalf("second connect status = %d, want %d", rec2.Code, http.StatusOK)
	}

	var conns []models.GmailConnection
	if err := s.db.Where("user_id = ?", s.userID).Find(&conns).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected exactly one connection row for the user, got %d", len(conns))
	}
	if conns[0].GmailAddress != "second@gmail.com" {
		t.Errorf("GmailAddress = %q, want the updated address %q", conns[0].GmailAddress, "second@gmail.com")
	}
}

func TestConnectGmail_NoRefreshTokenReturns400(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeExchangeGmailCode(t, &oauth2.Token{AccessToken: "access-1"}, "user@gmail.com", nil)
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{
		"code": "auth-code", "redirect_uri": "http://localhost:8081",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d when Google didn't grant offline access", rec.Code, http.StatusBadRequest)
	}
}

func TestConnectGmail_ExchangeErrorReturns400(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeExchangeGmailCode(t, nil, "", errors.New("invalid_grant"))
	defer restore()

	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{
		"code": "bad-code", "redirect_uri": "http://localhost:8081",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestConnectGmail_MissingFieldsReturns400(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestConnectGmail_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", "", map[string]any{
		"code": "x", "redirect_uri": "http://localhost:8081",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGmailStatus_NotConnectedByDefault(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodGet, "/gmail/status", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Connected bool `json:"connected"`
	}
	decodeJSON(t, rec, &body)
	if body.Connected {
		t.Error("Connected = true, want false when no connection exists")
	}
}

func TestGmailStatus_ConnectedAfterConnect(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeExchangeGmailCode(t, &oauth2.Token{AccessToken: "a1", RefreshToken: "r1", Expiry: time.Now().Add(time.Hour)}, "user@gmail.com", nil)
	defer restore()
	doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{"code": "c1", "redirect_uri": "http://localhost:8081"})

	rec := doJSONAs(t, s.router, http.MethodGet, "/gmail/status", s.token, nil)
	var body struct {
		Connected    bool   `json:"connected"`
		GmailAddress string `json:"gmail_address"`
	}
	decodeJSON(t, rec, &body)
	if !body.Connected {
		t.Error("Connected = false, want true after a successful connect")
	}
	if body.GmailAddress != "user@gmail.com" {
		t.Errorf("GmailAddress = %q, want %q", body.GmailAddress, "user@gmail.com")
	}
}

func TestDisconnectGmail_RemovesConnection(t *testing.T) {
	s := newTestSetup(t)
	restore := fakeExchangeGmailCode(t, &oauth2.Token{AccessToken: "a1", RefreshToken: "r1", Expiry: time.Now().Add(time.Hour)}, "user@gmail.com", nil)
	doJSONAs(t, s.router, http.MethodPost, "/gmail/connect", s.token, map[string]any{"code": "c1", "redirect_uri": "http://localhost:8081"})
	restore()

	originalRevoke := handlers.RevokeGoogleOAuthToken
	handlers.RevokeGoogleOAuthToken = func(context.Context, string) {}
	defer func() { handlers.RevokeGoogleOAuthToken = originalRevoke }()

	rec := doJSONAs(t, s.router, http.MethodDelete, "/gmail/disconnect", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var count int64
	s.db.Model(&models.GmailConnection{}).Where("user_id = ?", s.userID).Count(&count)
	if count != 0 {
		t.Errorf("expected the connection row to be deleted, found %d remaining", count)
	}

	statusRec := doJSONAs(t, s.router, http.MethodGet, "/gmail/status", s.token, nil)
	var body struct {
		Connected bool `json:"connected"`
	}
	decodeJSON(t, statusRec, &body)
	if body.Connected {
		t.Error("Connected = true after disconnect, want false")
	}
}

func TestDisconnectGmail_NoExistingConnectionStillSucceeds(t *testing.T) {
	s := newTestSetup(t)
	rec := doJSONAs(t, s.router, http.MethodDelete, "/gmail/disconnect", s.token, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
