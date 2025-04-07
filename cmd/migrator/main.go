package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wangfenjin/mojito/internal/app/database"
	"github.com/wangfenjin/mojito/internal/app/database/migrations"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/gorm"
)

func main() {
	config := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := database.NewConnection(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := generateMigrationSQL(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Migration SQL files generated successfully")
}

func generateMigrationSQL(db *gorm.DB) error {
	// Get all model versions
	modelVersions := models.GetModelVersions()
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
