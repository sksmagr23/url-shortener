package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource/mongo"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/handler"
	"github.com/sksmagr23/url-shortener-gofr/service"
	"github.com/sksmagr23/url-shortener-gofr/store"
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
	app.UseMiddleware(auth.JWTMiddleware(os.Getenv("JWT_SECRET")))

	// Health check endpoint
	app.GET("/health", handler.HealthHandler())

	userStore := store.NewUserStore()
	userService := service.NewUserService(userStore, os.Getenv("JWT_SECRET"))
	userHandler := handler.NewUserHandler(userService)

	urlStore := store.NewURLStore()
	shortURLHost := os.Getenv("SHORT_URL_HOST")
	urlService := service.NewURLService(urlStore, shortURLHost)
	urlHandler := handler.NewURLHandler(urlService)

	// User endpoints
	app.POST("/users/register", userHandler.Register)
	app.POST("/users/login", userHandler.Login)
	app.GET("/users/profile", userHandler.GetProfile)
	app.PUT("/users/profile", userHandler.UpdateProfile)
	app.POST("/users/api-key", userHandler.GenerateAPIKey)

	// URL endpoints
	app.POST("/urls", urlHandler.Create)
	app.GET("/urls/{short_code}", urlHandler.Get)
	app.GET("/{short_code}", urlHandler.Redirect)

	app.Run()
}
