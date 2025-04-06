package database

import (
	"fmt"
	"log"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/wangfenjin/mojito/internal/app/database/migrations"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Run migrations in development environment
	if os.Getenv("ENV") != "production" {
		if err := RunMigrations(db); err != nil {
			log.Printf("Failed to run migrations: %v", err)
			return nil, err
		}
		log.Printf("Migrations completed successfully")
	}

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	// Get all registered model versions
	registeredVersions := models.GetModelVersions()
	migrations := migrations.GenerateMigration(registeredVersions)
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.Migrate()
}
