package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"warrantykeeper/server/internal/auth"
)

const testSecret = "test-secret"

func TestGenerateAndParseAccessToken_RoundTrips(t *testing.T) {
	userID, householdID := uuid.New(), uuid.New()
	token, err := auth.GenerateAccessToken(testSecret, userID, householdID)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := auth.ParseAccessToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.HouseholdID != householdID {
		t.Errorf("HouseholdID = %v, want %v", claims.HouseholdID, householdID)
	}
}

func TestGenerateAndParseRefreshToken_RoundTrips(t *testing.T) {
	userID, householdID := uuid.New(), uuid.New()
	token, err := auth.GenerateRefreshToken(testSecret, userID, householdID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	claims, err := auth.ParseRefreshToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseRefreshToken returned error: %v", err)
	}
	if claims.UserID != userID || claims.HouseholdID != householdID {
		t.Errorf("claims = %+v, want UserID=%v HouseholdID=%v", claims, userID, householdID)
	}
}

func TestAccessTokenExpiresAfterTheDocumentedTTL(t *testing.T) {
	token, err := auth.GenerateAccessToken(testSecret, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	claims, err := auth.ParseAccessToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}

	wantExpiry := time.Now().Add(auth.AccessTokenTTL)
	diff := claims.ExpiresAt.Time.Sub(wantExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("ExpiresAt = %v, want within a minute of %v (now + AccessTokenTTL)", claims.ExpiresAt.Time, wantExpiry)
	}
}

func TestRefreshTokenTTLIsLongerThanAccessTokenTTL(t *testing.T) {
	if auth.RefreshTokenTTL <= auth.AccessTokenTTL {
		t.Errorf("RefreshTokenTTL (%v) should be longer than AccessTokenTTL (%v)", auth.RefreshTokenTTL, auth.AccessTokenTTL)
	}
}

func TestParseAccessToken_RejectsARefreshToken(t *testing.T) {
	refreshToken, err := auth.GenerateRefreshToken(testSecret, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if _, err := auth.ParseAccessToken(testSecret, refreshToken); err == nil {
		t.Error("ParseAccessToken accepted a refresh token, want an error")
	}
}

func TestParseRefreshToken_RejectsAnAccessToken(t *testing.T) {
	accessToken, err := auth.GenerateAccessToken(testSecret, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if _, err := auth.ParseRefreshToken(testSecret, accessToken); err == nil {
		t.Error("ParseRefreshToken accepted an access token, want an error")
	}
}

func TestParseAccessToken_RejectsWrongSecret(t *testing.T) {
	token, err := auth.GenerateAccessToken(testSecret, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if _, err := auth.ParseAccessToken("a-different-secret", token); err == nil {
		t.Error("ParseAccessToken accepted a token signed with a different secret, want an error")
	}
}

func TestParseAccessToken_RejectsGarbage(t *testing.T) {
	if _, err := auth.ParseAccessToken(testSecret, "not-a-real-token"); err == nil {
		t.Error("ParseAccessToken accepted garbage input, want an error")
	}
	if _, err := auth.ParseAccessToken(testSecret, ""); err == nil {
		t.Error("ParseAccessToken accepted an empty string, want an error")
	}
}

func TestParseAccessToken_RejectsExpiredToken(t *testing.T) {
	// GenerateAccessToken only offers the fixed real TTL, so exercise the
	// underlying jwt library the same way auth.go does to produce an
	// already-expired token signed the same way.
	claims := auth.Claims{
		UserID:      uuid.New(),
		HouseholdID: uuid.New(),
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	if _, err := auth.ParseAccessToken(testSecret, signed); err == nil {
		t.Error("ParseAccessToken accepted an expired token, want an error")
	}
}
