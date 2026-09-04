package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/auth"
)

const (
	CtxUserID      = "user_id"
	CtxHouseholdID = "household_id"
)

// RequireAuth validates the Bearer access token and puts user_id /
// household_id on the Gin context for downstream handlers to scope
// queries by (see architecture doc section 6 on application-layer
// multi-tenancy).
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		claims, err := auth.ParseAccessToken(jwtSecret, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxHouseholdID, claims.HouseholdID)
		c.Next()
	}
}
