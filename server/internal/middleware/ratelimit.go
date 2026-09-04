package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit throttles requests per client IP. Applied to auth and OCR-facing
// endpoints per the architecture doc's security section ("Rate limiting: על
// endpoints של OCR ו-auth למניעת שימוש לרעה"). An in-memory limiter is
// sufficient for a single-instance MVP deployment.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(key string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := limiters[key]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			limiters[key] = l
		}
		return l
	}

	return func(c *gin.Context) {
		if !getLimiter(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, slow down"})
			return
		}
		c.Next()
	}
}
