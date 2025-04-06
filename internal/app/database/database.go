package database

import (
	"fmt"
	"log"

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

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Item{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return nil, err
	}

	return db, nil
}
