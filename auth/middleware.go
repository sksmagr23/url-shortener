package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type APIKeyValidator interface {
	FindByAPIKey(ctx *gofr.Context, apiKey string) (*model.User, error)
}

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return AuthMiddleware(secret, nil)
}

func AuthMiddleware(secret string, validator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r.Header.Get("Authorization"))
			apiKeyHeader := strings.TrimSpace(r.Header.Get("X-API-Key"))

			// Check if API key is supplied via X-API-Key or Authorization Bearer usk_...
			var apiKey string
			if apiKeyHeader != "" {
				apiKey = apiKeyHeader
			} else if strings.HasPrefix(tokenString, "usk_") {
				apiKey = tokenString
			}

			if !requiresAuth(r) {
				if apiKey != "" && validator != nil {
					gofrCtx := &gofr.Context{Context: r.Context()}
					user, err := validator.FindByAPIKey(gofrCtx, apiKey)
					if err == nil && user != nil {
						r = r.WithContext(ContextWithUserID(r.Context(), user.ID))
					}
				} else if tokenString != "" && !strings.HasPrefix(tokenString, "usk_") {
					claims, err := ValidateToken(tokenString, secret)
					if err == nil && claims != nil {
						r = r.WithContext(ContextWithUserID(r.Context(), claims.UserID))
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// Required authentication endpoints
			if apiKey != "" && validator != nil {
				gofrCtx := &gofr.Context{Context: r.Context()}
				user, err := validator.FindByAPIKey(gofrCtx, apiKey)
				if err != nil || user == nil {
					writeUnauthorized(w, "invalid API key")
					return
				}
				ctx := ContextWithUserID(r.Context(), user.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if tokenString == "" {
				writeUnauthorized(w, "missing authentication token or API key")
				return
			}

			if strings.HasPrefix(tokenString, "usk_") {
				writeUnauthorized(w, "invalid API key")
				return
			}

			claims, err := ValidateToken(tokenString, secret)
			if err != nil {
				writeUnauthorized(w, "invalid bearer token")
				return
			}

			ctx := ContextWithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requiresAuth(r *http.Request) bool {
	path := r.URL.Path

	if path == "/users/register" || path == "/users/login" || path == "/urls/public" {
		return false
	}

	return strings.HasPrefix(path, "/urls") ||
		path == "/users/profile" ||
		path == "/users/api-key" ||
		strings.HasPrefix(path, "/users/api-keys")
}

func bearerToken(header string) string {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}

	return strings.TrimSpace(token)
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"message": message,
		},
	})
}
