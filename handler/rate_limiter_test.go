package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sksmagr23/url-shortener-gofr/handler"
)

func TestRateLimiterMiddlewareEnforcesLimit(t *testing.T) {
	limiter := handler.RateLimiterMiddleware(2) // limit 2 requests per minute

	testHandler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request 1: Allowed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "203.0.113.1:1234"
	rr1 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Request 2: Allowed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "203.0.113.1:1234"
	rr2 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// Request 3: Exceeded -> HTTP 429
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "203.0.113.1:1234"
	rr3 := httptest.NewRecorder()
	testHandler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusTooManyRequests, rr3.Code)
	assert.Equal(t, "60", rr3.Header().Get("Retry-After"))
}
