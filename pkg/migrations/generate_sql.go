package migrations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func GenerateSQL(db *gorm.DB, modelVersions []ModelVersion) (map[string]string, error) {
	// Get all model versions
	migrationList := GenerateMigration(modelVersions)
	if len(migrationList) == 0 {
		return nil, fmt.Errorf("no migrations generated")
	}

	sqls := make(map[string]string)
	// Generate SQL for each migration
	for _, m := range migrationList {
		// up
		sql, err := generateSQL(db, m, false)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s_up.sql", m.ID)
		sqls[filename] = sql

		// down
		sql, err = generateSQL(db, m, true)
		if err != nil {
			return nil, err
		}

		filename = fmt.Sprintf("%s_down.sql", m.ID)
		sqls[filename] = sql
	}

	return sqls, nil
}

func generateSQL(db *gorm.DB, m *gormigrate.Migration, isRollback bool) (string, error) {
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
	if isRollback {
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
