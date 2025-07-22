package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/datasource/mongo"

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

	// Health check endpoint
	app.GET("/health", handler.HealthHandler())

	urlStore := store.NewURLStore()
	shortURLHost := os.Getenv("SHORT_URL_HOST")
	urlService := service.NewURLService(urlStore, shortURLHost)
	urlHandler := handler.NewURLHandler(urlService)

	userStore := store.NewUserStore()
	jwtSecret := os.Getenv("JWT_SECRET")
	userService := service.NewUserService(userStore, jwtSecret)
	userHandler := handler.NewUserHandler(userService)

	// User endpoints
	app.POST("/users/register", userHandler.Register)
	app.POST("/users/login", userHandler.Login)
	app.GET("/public-test", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]string{"msg": "public"}, nil
	})

	app.EnableOAuth("https://www.googleapis.com/oauth2/v3/certs", 60, jwt.WithExpirationRequired())

	app.GET("/users/profile", userHandler.Profile)
	app.PUT("/users/profile", userHandler.UpdateProfile)
	app.POST("/users/api-key", userHandler.GenerateAPIKey)

	// URL endpoints
	app.POST("/urls", urlHandler.Create)
	app.GET("/urls/{short_code}", urlHandler.Get)
	app.GET("/{short_code}", urlHandler.Redirect)

	app.Run()
}
