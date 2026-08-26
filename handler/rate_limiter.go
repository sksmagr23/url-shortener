package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/service"
)

type memoryLimiter struct {
	mu       sync.Mutex
	counts   map[string]int
	resetAt map[string]time.Time
}

var memLimiter = &memoryLimiter{
	counts:   make(map[string]int),
	resetAt: make(map[string]time.Time),
}

// RateLimiterMiddleware enforces a sliding window rate limit per client identifier (IP or User ID).
func RateLimiterMiddleware(limit int) func(http.Handler) http.Handler {
	if limit <= 0 {
		limit = 100 // Default: 100 requests per minute
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Identify client by User ID if authenticated, else IP address
			identifier := service.ExtractIP(r)
			if userID, ok := auth.UserIDFromContext(r.Context()); ok && userID != "" {
				identifier = "user:" + userID
			}

			now := time.Now()
			minuteKey := fmt.Sprintf("%s:%d", identifier, now.Unix()/60)

			memLimiter.mu.Lock()
			// Clean up old window entries
			for k, reset := range memLimiter.resetAt {
				if now.After(reset) {
					delete(memLimiter.counts, k)
					delete(memLimiter.resetAt, k)
				}
			}

			count := memLimiter.counts[minuteKey] + 1
			memLimiter.counts[minuteKey] = count
			if count == 1 {
				memLimiter.resetAt[minuteKey] = now.Add(time.Minute)
			}
			memLimiter.mu.Unlock()

			if count > limit {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    http.StatusTooManyRequests,
						"message": "rate limit exceeded, please try again later",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
