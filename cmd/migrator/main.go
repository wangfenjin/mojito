package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
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
	migrations := migrations.GenerateMigration(modelVersions)

	// Create scripts/db directory if not exists
	if err := os.MkdirAll("scripts/db", 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate SQL for each migration
	for _, m := range migrations {
		sql, err := getMigrationSQL(db, m)
		if err != nil {
			return err
		}
		fmt.Printf("Generated SQL for migration %s:\n%s\n", m.ID, sql)

		filename := fmt.Sprintf("scripts/db/%s_%s.sql", m.ID, time.Now().Format("20060102150405"))
		if err := os.WriteFile(filename, []byte(sql), 0644); err != nil {
			return fmt.Errorf("failed to write migration file: %w", err)
		}
	}

	return nil
}

func getMigrationSQL(db *gorm.DB, m *gormigrate.Migration) (string, error) {
	var statements []string
	// Create a new session with DryRun mode
	dryDB := db.Session(&gorm.Session{
		DryRun:      true,
		PrepareStmt: false,
	})

	// Run migration
	if err := m.Migrate(dryDB.Debug()); err != nil {
		return "", fmt.Errorf("failed to generate SQL for migration %s: %w", m.ID, err)
	}

	// TODO: how to get the SQL statements?

	// // Combine all SQL statements
	// if len(statements) == 0 {
	// 	return "", fmt.Errorf("no SQL statements generated for migration %s", m.ID)
	// }

	return "-- Migration: " + m.ID + "\n" +
		"BEGIN;\n" +
		join(statements, ";\n") +
		";\nCOMMIT;\n", nil
}

func join(statements []string, sep string) string {
	result := ""
	for i, stmt := range statements {
		if stmt != "" {
			if i > 0 {
				result += sep
			}
			result += stmt
		}
	}
	return result
}
