package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 24 * time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour

	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	HouseholdID uuid.UUID `json:"household_id"`
	TokenType   string    `json:"token_type"`
	jwt.RegisteredClaims
}

func generateToken(secret string, userID, householdID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:      userID,
		HouseholdID: householdID,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateAccessToken(secret string, userID, householdID uuid.UUID) (string, error) {
	return generateToken(secret, userID, householdID, tokenTypeAccess, AccessTokenTTL)
}

func GenerateRefreshToken(secret string, userID, householdID uuid.UUID) (string, error) {
	return generateToken(secret, userID, householdID, tokenTypeRefresh, RefreshTokenTTL)
}

func parseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ParseAccessToken(secret, tokenStr string) (*Claims, error) {
	claims, err := parseToken(secret, tokenStr)
	if err != nil || claims.TokenType != tokenTypeAccess {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ParseRefreshToken(secret, tokenStr string) (*Claims, error) {
	claims, err := parseToken(secret, tokenStr)
	if err != nil || claims.TokenType != tokenTypeRefresh {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
