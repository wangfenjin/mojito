package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wangfenjin/mojito/internal/app/database"
	"github.com/wangfenjin/mojito/internal/app/middleware"
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
