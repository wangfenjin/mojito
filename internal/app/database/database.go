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
type ConnectionParams struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// Connect creates a new database connection
func Connect(params ConnectionParams) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		params.Host, params.Port, params.User, params.Password, params.DBName, params.SSLMode, params.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
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
