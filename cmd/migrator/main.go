package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wangfenjin/mojito/internal/app/config"
	"github.com/wangfenjin/mojito/internal/app/database"

	// Import models to register them with gorm
	_ "github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/pkg/migrations"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.Connect(database.ConnectionParams{
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

	if err := generateMigrationSQL(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Migration SQL files generated successfully")
}

func generateMigrationSQL(db *gorm.DB) error {
	// Get all model versions
	modelVersions := migrations.GetModelVersions()
	if len(modelVersions) == 0 {
		return fmt.Errorf("no model versions found")
	}

	// Create scripts/db directory if not exists
	if err := os.MkdirAll("scripts/db", 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate SQLs
	sqls, err := migrations.GenerateSQL(db, modelVersions)
	if err != nil {
		return err
	}
	for filename, sql := range sqls {
		if err := os.WriteFile(fmt.Sprintf("scripts/db/%s", filename), []byte(sql), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}
