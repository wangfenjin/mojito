package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/wangfenjin/mojito/internal/app/database"
	"github.com/wangfenjin/mojito/internal/app/database/migrations"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
		// up
		sql, err := getMigrationSQL(db, m, false)
		if err != nil {
			return err
		}
		fmt.Printf("Generated SQL for migration %s:\n%s\n", m.ID, sql)

		filename := fmt.Sprintf("scripts/db/%s_up.sql", m.ID)
		if err := os.WriteFile(filename, []byte(sql), 0644); err != nil {
			return fmt.Errorf("failed to write up file: %w", err)
		}

		// down
		sql, err = getMigrationSQL(db, m, true)
		if err != nil {
			return err
		}
		fmt.Printf("Generated SQL for rollback %s:\n%s\n", m.ID, sql)

		filename = fmt.Sprintf("scripts/db/%s_down.sql", m.ID)
		if err := os.WriteFile(filename, []byte(sql), 0644); err != nil {
			return fmt.Errorf("failed to write down file: %w", err)
		}
	}

	return nil
}

// Add this type at the top of the file after imports
type SQLLogger struct {
	Statements []string
	LogLevel   logger.LogLevel
}

func (l *SQLLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *SQLLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	// We only care about SQL statements, so this is empty
}

func (l *SQLLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	// We only care about SQL statements, so this is empty
}

func (l *SQLLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// We only care about SQL statements, so this is empty
}

func (l *SQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	if sql != "" && !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT") {
		l.Statements = append(l.Statements, sql)
	}
}

// Update getMigrationSQL to initialize the logger properly
func getMigrationSQL(db *gorm.DB, m *gormigrate.Migration, is_rollback bool) (string, error) {
	logger := &SQLLogger{
		Statements: make([]string, 0),
		LogLevel:   logger.Info, // Set to Info to capture all SQL statements
	}

	// Create a new session with DryRun mode and custom logger
	dryDB := db.Session(&gorm.Session{
		DryRun:      true,
		PrepareStmt: false,
		Logger:      logger,
	})

	// Run migration
	if is_rollback {
		if err := m.Rollback(dryDB); err != nil {
			return "", fmt.Errorf("failed to generate SQL for rollback migration %s: %w", m.ID, err)
		}
	} else {
		if err := m.Migrate(dryDB); err != nil {
			return "", fmt.Errorf("failed to generate SQL for migration %s: %w", m.ID, err)
		}
	}

	// Combine all SQL statements
	if len(logger.Statements) == 0 {
		return "", fmt.Errorf("no SQL statements generated for migration %s", m.ID)
	}

	return "-- Migration: " + m.ID + "\n" +
		"BEGIN;\n" +
		join(logger.Statements, ";\n") +
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
