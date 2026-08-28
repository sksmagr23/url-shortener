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

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "url_shortener"
	}

	db := mongo.New(mongo.Config{
		URI:               mongoURI,
		Database:          mongoDB,
		ConnectionTimeout: 4 * time.Second,
	})

	app.AddMongo(db)
	app.UseMiddleware(handler.MetadataMiddleware())
	app.UseMiddleware(handler.RateLimiterMiddleware(100))

	userStore := store.NewUserStoreWithMongo(db)
	app.UseMiddleware(auth.AuthMiddleware(os.Getenv("JWT_SECRET"), userStore))

	// Health & status endpoints
	app.GET("/health", handler.HealthHandler())
	app.GET("/api/status", handler.StatusHandler())

	userService := service.NewUserService(userStore, os.Getenv("JWT_SECRET"))
	userHandler := handler.NewUserHandler(userService)

	urlStore := store.NewURLStoreWithMongo(db)
	analyticsStore := store.NewAnalyticsStoreWithMongo(db)
	urlCache := store.NewURLCache()
	shortURLHost := os.Getenv("SHORT_URL_HOST")
	urlService := service.NewURLService(urlStore, analyticsStore, urlCache, shortURLHost)
	urlHandler := handler.NewURLHandler(urlService)

	// User endpoints
	app.POST("/users/register", userHandler.Register)
	app.POST("/users/login", userHandler.Login)
	app.GET("/users/profile", userHandler.GetProfile)
	app.PUT("/users/profile", userHandler.UpdateProfile)
	app.POST("/users/api-key", userHandler.GenerateAPIKey)
	app.GET("/users/api-keys", userHandler.ListAPIKeys)
	app.DELETE("/users/api-keys/{api_key}", userHandler.RevokeAPIKey)

	// URL endpoints
	app.POST("/urls", urlHandler.Create)
	app.GET("/urls", urlHandler.List)
	app.GET("/urls/public", urlHandler.ListPublic)
	app.GET("/urls/{short_code}/qr", urlHandler.GetQRCode)
	app.GET("/urls/{short_code}", urlHandler.Get)
	app.GET("/urls/{short_code}/analytics", urlHandler.GetAnalyticsSummary)
	app.GET("/urls/{short_code}/analytics/timeseries", urlHandler.GetAnalyticsTimeseries)
	app.PUT("/urls/{short_code}", urlHandler.Update)
	app.DELETE("/urls/{short_code}", urlHandler.Delete)
	app.GET("/{short_code}", urlHandler.Redirect)

	app.Run()
}
