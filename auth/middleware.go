package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r.Header.Get("Authorization"))

			if !requiresJWT(r) {
				if tokenString != "" {
					claims, err := ValidateToken(tokenString, secret)
					if err != nil {
						writeUnauthorized(w, "invalid bearer token")

						return
					}

					r = r.WithContext(ContextWithUserID(r.Context(), claims.UserID))
				}
				next.ServeHTTP(w, r)

				return
			}

			if tokenString == "" {
				writeUnauthorized(w, "missing bearer token")

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

func requiresJWT(r *http.Request) bool {
	path := r.URL.Path

	if path == "/users/register" || path == "/users/login" {
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
