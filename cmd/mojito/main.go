package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	"github.com/wangfenjin/mojito/internal/app/database"
	"github.com/wangfenjin/mojito/internal/app/repository"
	"github.com/wangfenjin/mojito/internal/app/routes"
)

func main() {
	// Initialize database connection
	dbConfig := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "mojito",
		SSLMode:  "disable",
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)

	// Create Hertz server
	h := server.Default()

	// Add CORS middleware
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Add middleware to inject repositories into context
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		ctx = context.WithValue(ctx, "userRepository", userRepo)
		ctx = context.WithValue(ctx, "itemRepository", itemRepo)
		c.Next(ctx)
	})

	// Set up API routes
	routes.RegisterRoutes(h)
	if os.Getenv("ENV") != "production" {
		routes.RegisterTestRoutes(h)
	}

	// Start the server
	h.Spin()
}
