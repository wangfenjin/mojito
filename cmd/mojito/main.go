// Package main is the entry point for the Mojito HTTP server
package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wangfenjin/mojito/internal/app/config"
	"github.com/wangfenjin/mojito/internal/app/database"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/repository"
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
	db, err := database.Connect(database.ConnectionParams{
		Type:       cfg.Database.Type,
		Host:       cfg.Database.Host,
		Port:       cfg.Database.Port,
		User:       cfg.Database.User,
		Password:   cfg.Database.Password,
		DBName:     cfg.Database.Name,
		SSLMode:    cfg.Database.SSLMode,
		TimeZone:   cfg.Database.TimeZone,
		SQLitePath: cfg.Database.SQLitePath,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)

	// Create Hertz server
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(middleware.LoggerMiddleware())

	// Add middleware to inject repositories into context
	r.Use(func(c *gin.Context) {
		c.Set("userRepository", userRepo)
		c.Set("itemRepository", itemRepo)
		c.Next()
	})

	// Set up API routes
	routes.RegisterRoutes(r)
	if os.Getenv("ENV") != "production" {
		routes.RegisterTestRoutes(r)
	}

	// Start the server
	if err := r.Run(); err != nil {
		panic(err)
	}
}
