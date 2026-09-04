package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authpkg "warrantykeeper/server/internal/auth"
	"warrantykeeper/server/internal/middleware"
)

const testSecret = "test-secret"

// newRouter wires RequireAuth in front of a terminal handler that records
// whether it was reached and echoes back the context values RequireAuth is
// supposed to set, so tests can assert on both "did it let the request
// through" and "did it set the right identity".
func newRouter(secret string) (router *gin.Engine, reached *bool, gotUserID, gotHouseholdID *uuid.UUID) {
	gin.SetMode(gin.TestMode)
	reached = new(bool)
	gotUserID = new(uuid.UUID)
	gotHouseholdID = new(uuid.UUID)

	router = gin.New()
	router.GET("/protected", middleware.RequireAuth(secret), func(c *gin.Context) {
		*reached = true
		*gotUserID = c.MustGet(middleware.CtxUserID).(uuid.UUID)
		*gotHouseholdID = c.MustGet(middleware.CtxHouseholdID).(uuid.UUID)
		c.Status(http.StatusOK)
	})
	return
}

func doRequest(router *gin.Engine, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRequireAuth_ValidToken_SetsContextAndCallsNext(t *testing.T) {
	router, reached, gotUserID, gotHouseholdID := newRouter(testSecret)
	userID, householdID := uuid.New(), uuid.New()
	token, err := authpkg.GenerateAccessToken(testSecret, userID, householdID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	rec := doRequest(router, "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !*reached {
		t.Fatal("expected the downstream handler to be called")
	}
	if *gotUserID != userID {
		t.Errorf("context user_id = %v, want %v", *gotUserID, userID)
	}
	if *gotHouseholdID != householdID {
		t.Errorf("context household_id = %v, want %v", *gotHouseholdID, householdID)
	}
}

func TestRequireAuth_BearerSchemeIsCaseInsensitive(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)
	token, _ := authpkg.GenerateAccessToken(testSecret, uuid.New(), uuid.New())

	rec := doRequest(router, "bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !*reached {
		t.Error("expected the downstream handler to be called for a lowercase \"bearer\" scheme")
	}
}

func TestRequireAuth_MissingHeaderIsRejected(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)
	rec := doRequest(router, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if *reached {
		t.Error("downstream handler must not run without an Authorization header")
	}
}

func TestRequireAuth_MalformedHeaderIsRejected(t *testing.T) {
	token, _ := authpkg.GenerateAccessToken(testSecret, uuid.New(), uuid.New())
	tests := []struct {
		name   string
		header string
	}{
		{"wrong scheme", "Token " + token},
		{"missing the token part entirely", "Bearer"},
		{"missing the Bearer prefix", token},
		{"blank token", "Bearer  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, reached, _, _ := newRouter(testSecret)
			rec := doRequest(router, tt.header)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			if *reached {
				t.Error("downstream handler must not run for a malformed header")
			}
		})
	}
}

func TestRequireAuth_GarbageTokenIsRejected(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)
	rec := doRequest(router, "Bearer not-a-real-token")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if *reached {
		t.Error("downstream handler must not run for an invalid token")
	}
}

func TestRequireAuth_TokenSignedWithWrongSecretIsRejected(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)
	token, _ := authpkg.GenerateAccessToken("a-completely-different-secret", uuid.New(), uuid.New())

	rec := doRequest(router, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if *reached {
		t.Error("downstream handler must not run for a token signed with a different secret")
	}
}

func TestRequireAuth_RefreshTokenIsRejected(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)
	// A refresh token must not work as an access token against a protected route.
	refreshToken, err := authpkg.GenerateRefreshToken(testSecret, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	rec := doRequest(router, "Bearer "+refreshToken)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if *reached {
		t.Error("downstream handler must not run for a refresh token")
	}
}

func TestRequireAuth_ExpiredTokenIsRejected(t *testing.T) {
	router, reached, _, _ := newRouter(testSecret)

	// Hand-craft an already-expired access token with the same shape
	// auth.GenerateAccessToken produces, since that helper only offers a
	// fixed TTL.
	claims := struct {
		UserID      uuid.UUID `json:"user_id"`
		HouseholdID uuid.UUID `json:"household_id"`
		TokenType   string    `json:"token_type"`
		jwt.RegisteredClaims
	}{
		UserID:      uuid.New(),
		HouseholdID: uuid.New(),
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	rec := doRequest(router, "Bearer "+signed)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if *reached {
		t.Error("downstream handler must not run for an expired token")
	}
}
