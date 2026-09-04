package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/middleware"
)

func newRateLimitRouter(rps float64, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/limited", middleware.RateLimit(rps, burst), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func requestFrom(router *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRateLimit_AllowsUpToBurstThenBlocks(t *testing.T) {
	router := newRateLimitRouter(1, 3) // 1 req/s sustained, burst of 3

	for i := 1; i <= 3; i++ {
		rec := requestFrom(router, "192.0.2.1:1111")
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d (burst should allow it)", i, rec.Code, http.StatusOK)
		}
	}

	rec := requestFrom(router, "192.0.2.1:1111")
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: status = %d, want %d (burst exhausted)", rec.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimit_TracksEachClientIPSeparately(t *testing.T) {
	router := newRateLimitRouter(1, 1) // burst of 1 — the 2nd request from the same IP always blocks

	recA1 := requestFrom(router, "192.0.2.1:1111")
	if recA1.Code != http.StatusOK {
		t.Fatalf("client A, request 1: status = %d, want %d", recA1.Code, http.StatusOK)
	}
	recA2 := requestFrom(router, "192.0.2.1:2222") // same IP, different port — ClientIP() ignores the port
	if recA2.Code != http.StatusTooManyRequests {
		t.Errorf("client A, request 2: status = %d, want %d", recA2.Code, http.StatusTooManyRequests)
	}

	// A different client IP must have its own, independent budget.
	recB1 := requestFrom(router, "192.0.2.2:1111")
	if recB1.Code != http.StatusOK {
		t.Errorf("client B, request 1: status = %d, want %d (must not be blocked by client A's usage)", recB1.Code, http.StatusOK)
	}
}

func TestRateLimit_RefillsOverTime(t *testing.T) {
	router := newRateLimitRouter(1000, 1) // burst of 1, refilling ~1 token/ms

	rec1 := requestFrom(router, "192.0.2.3:1111")
	if rec1.Code != http.StatusOK {
		t.Fatalf("request 1: status = %d, want %d", rec1.Code, http.StatusOK)
	}
	rec2 := requestFrom(router, "192.0.2.3:1111")
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("request 2 (immediate): status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}

	time.Sleep(20 * time.Millisecond) // several multiples of the ~1ms refill interval

	rec3 := requestFrom(router, "192.0.2.3:1111")
	if rec3.Code != http.StatusOK {
		t.Errorf("request 3 (after waiting): status = %d, want %d (bucket should have refilled)", rec3.Code, http.StatusOK)
	}
}
