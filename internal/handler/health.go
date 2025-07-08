package handler

import (
	"time"

	"gofr.dev/pkg/gofr"
)

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
