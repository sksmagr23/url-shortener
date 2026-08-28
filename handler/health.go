package handler

import (
	"time"

	"gofr.dev/pkg/gofr"
)

// GET /health
func HealthHandler() func(ctx *gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		mongoStatus := "connected"
		_, err := ctx.Mongo.CountDocuments(ctx, "urls", map[string]interface{}{})
		if err != nil {
			mongoStatus = "disconnected"
		}
		return map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": map[string]string{
				"mongoDB": mongoStatus,
			},
		}, nil
	}
}

// GET /api/status
func StatusHandler() func(ctx *gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		totalURLs, _ := ctx.Mongo.CountDocuments(ctx, "urls", map[string]interface{}{})
		totalClicks, _ := ctx.Mongo.CountDocuments(ctx, "click_events", map[string]interface{}{})
		totalUsers, _ := ctx.Mongo.CountDocuments(ctx, "users", map[string]interface{}{})

		return map[string]interface{}{
			"version":      "2.0.0",
			"status":       "OPERATIONAL",
			"total_urls":   totalURLs,
			"total_clicks": totalClicks,
			"total_users":  totalUsers,
		}, nil
	}
}
