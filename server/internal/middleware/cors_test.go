package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/middleware"
)

func newCORSRouter() (router *gin.Engine, reached *bool) {
	gin.SetMode(gin.TestMode)
	reached = new(bool)
	router = gin.New()
	router.Use(middleware.CORS())
	router.Any("/anything", func(c *gin.Context) {
		*reached = true
		c.Status(http.StatusOK)
	})
	return
}

func assertCORSHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected a non-empty Access-Control-Allow-Methods header")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected a non-empty Access-Control-Allow-Headers header")
	}
}

func TestCORS_NormalRequestGetsHeadersAndReachesHandler(t *testing.T) {
	router, reached := newCORSRouter()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !*reached {
		t.Error("expected the downstream handler to run for a normal GET")
	}
	assertCORSHeaders(t, rec)
}

func TestCORS_PreflightOPTIONSShortCircuits(t *testing.T) {
	router, reached := newCORSRouter()
	req := httptest.NewRequest(http.MethodOptions, "/anything", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if *reached {
		t.Error("an OPTIONS preflight must not reach the downstream handler")
	}
	assertCORSHeaders(t, rec)
}
