package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sksmagr23/url-shortener-gofr/handler"
	"github.com/sksmagr23/url-shortener-gofr/service"
	"github.com/sksmagr23/url-shortener-gofr/store"
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
	app.GET("/health", handler.HealthHandler())

	urlStore := store.NewURLStore()
	urlService := service.NewURLService(urlStore)
	urlHandler := handler.NewURLHandler(urlService)

	// URL endpoints
	app.POST("/urls", urlHandler.Create)
	app.GET("/urls/{short_code}", urlHandler.Get)
	app.GET("/{short_code}", urlHandler.Redirect)

	app.Run()
}
