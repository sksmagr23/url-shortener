package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sksmagr23/url-shortener-gofr/internal/handler"
	"github.com/sksmagr23/url-shortener-gofr/internal/service"
	"github.com/sksmagr23/url-shortener-gofr/internal/store"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource/mongo"
)

func main() {
	err := godotenv.Load("./configs/.env")
	if err != nil {
		fmt.Println("Error loading .env config:", err)
	}
	
	app := gofr.New()
	
	db := mongo.New(mongo.Config{
		URI:               os.Getenv("MONGO_URI"),
		Database:          os.Getenv("MONGO_DB"),
		ConnectionTimeout: 4 * time.Second,
	})

	app.AddMongo(db)

	// Health check endpoint
	app.GET("/health", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]interface{}{
			"status": "healthy",
			"services": map[string]string{
				"mongoDB": "connected",
			},
		}, nil
	})

	urlStore := store.NewURLStore()
	urlService := service.NewURLService(urlStore)
	urlHandler := handler.NewURLHandler(urlService)
	
	// URL endpoints
	app.POST("/api/urls", urlHandler.Create)
	app.GET("/api/urls/{short_code}", urlHandler.Get)

	app.Run()
}
