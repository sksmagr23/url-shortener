package handler

import (
	"net/http"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/service"
)

// MetadataMiddleware extracts IP, User-Agent, Referrer, and parsed device/geo details,
// embedding them into the request context for downstream services.
func MetadataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := service.ExtractIP(r)
			country := service.ResolveCountry(ip)
			ua := r.Header.Get("User-Agent")
			uaInfo := service.ParseUserAgent(ua)
			referer := r.Header.Get("Referer")

			meta := auth.RequestMetadata{
				IPAddress:  ip,
				UserAgent:  ua,
				Referrer:   referer,
				Browser:    uaInfo.Browser,
				OS:         uaInfo.OS,
				DeviceType: uaInfo.DeviceType,
				Country:    country,
			}

			ctx := auth.ContextWithMetadata(r.Context(), meta)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
