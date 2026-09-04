package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS is a permissive dev-only policy so the Expo web preview (served from
// its own localhost port) can call the API. Tighten to an explicit origin
// allowlist before any real deployment.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
