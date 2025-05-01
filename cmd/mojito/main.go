// Package main is the entry point for the Mojito HTTP server
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/wangfenjin/mojito/internal/app/config"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/routes"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.GetLogger().Error("Failed to load configuration", "err", err)
		panic(err)
	}
	logger.GetLogger().Info("Configuration loaded", "config", cfg)

	// Initialize database connection
	_, err = models.Connect(models.ConnectionParams{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
		TimeZone: cfg.Database.TimeZone,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create Chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Set up API routes
	routes.RegisterRoutes(r)
	if os.Getenv("ENV") != "production" {
		routes.RegisterTestRoutes(r)
	}

	// Start the server
	port := strconv.Itoa(cfg.Server.Port)
	logger.GetLogger().Info("Starting server on :" + port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		panic(err)
	}
}
