// Package database provides database connection and migration functionality
package database

import (
	"fmt"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
	"github.com/wangfenjin/mojito/pkg/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConnectionParams holds the parameters for connecting to the database
type ConnectionParams struct {
	Type       string
	Host       string
	Port       int
	User       string
	Password   string
	DBName     string
	SSLMode    string
	TimeZone   string
	SQLitePath string
}

// Connect establishes a connection to the database
func Connect(params ConnectionParams) (*gorm.DB, error) {
	db, err := connect(params)
	if err != nil {
		return nil, err
	}
	// Run migrations in development environment
	if os.Getenv("ENV") != "production" {
		if err := RunMigrations(db); err != nil {
			logger.GetLogger().Error("Failed to run migrations", "err", err)
			return nil, err
		}
	}
	return db, nil
}
func connect(params ConnectionParams) (*gorm.DB, error) {
	// Check environment - use SQLite for local development
	if params.Type == "sqlite3" {
		return connectSQLite(params.SQLitePath)
	}

	// Default to PostgreSQL
	return connectPostgres(params)
}

// connectPostgres establishes a connection to a PostgreSQL database
func connectPostgres(params ConnectionParams) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		params.Host, params.Port, params.User, params.Password, params.DBName, params.SSLMode, params.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	return db, nil
}

// connectSQLite establishes a connection to a SQLite database
func connectSQLite(dbPath string) (*gorm.DB, error) {
	// Ensure directory exists
	dir := dbPath[:len(dbPath)-len("/mojito.db")]
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for SQLite database: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	return db, nil
}

// RunMigrations executes database migrations for all registered models
func RunMigrations(db *gorm.DB) error {
	// Get all registered model versions
	migrations := migrations.GenerateMigration(migrations.GetModelVersions())
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.Migrate()
}
