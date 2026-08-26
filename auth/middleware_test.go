package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/model"
)

type mockValidator struct {
	user *model.User
	err  error
}

func (m *mockValidator) FindByAPIKey(ctx *gofr.Context, apiKey string) (*model.User, error) {
	if apiKey == "usk_valid_key" {
		return &model.User{ID: "user-123", Username: "apikeyuser"}, nil
	}
	return nil, m.err
}

func TestAuthMiddlewareWithAPIKey(t *testing.T) {
	validator := &mockValidator{}
	middleware := auth.AuthMiddleware("secret", validator)

	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)
		w.WriteHeader(http.StatusOK)
	}))

	// Request with valid X-API-Key
	req := httptest.NewRequest("GET", "/urls", nil)
	req.Header.Set("X-API-Key", "usk_valid_key")
	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddlewareWithInvalidAPIKey(t *testing.T) {
	validator := &mockValidator{}
	middleware := auth.AuthMiddleware("secret", validator)

	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with invalid X-API-Key
	req := httptest.NewRequest("GET", "/urls", nil)
	req.Header.Set("X-API-Key", "usk_invalid")
	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
