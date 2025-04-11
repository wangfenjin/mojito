package database

import (
	"fmt"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
	"github.com/wangfenjin/mojito/pkg/migrations"
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
			logger.GetLogger().Error("Failed to run migrations", "err", err)
			return nil, err
		}
		logger.GetLogger().Info("Migrations completed successfully")
	}

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	// Get all registered model versions
	migrations := migrations.GenerateMigration(migrations.GetModelVersions())
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.Migrate()
}
