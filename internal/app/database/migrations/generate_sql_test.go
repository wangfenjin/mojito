package migrations

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Base model for testing
type TestModel struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"not null"`
}

func TestGenerateSQL(t *testing.T) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_PORT", "5432"),
		getEnvOrDefault("DB_USER", "postgres"),
		getEnvOrDefault("DB_PASSWORD", "postgres"),
		getEnvOrDefault("DB_NAME", "mojito"),
		getEnvOrDefault("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	tests := []struct {
		name     string
		skip     bool
		versions []models.ModelVersion
		wantSQLs map[string][]string
		wantErr  bool
	}{
		{
			name: "create new table",
			versions: []models.ModelVersion{
				{
					Version: "1.0.0",
					Current: &struct {
						TestModel
						Name  string `gorm:"size:100;not null"`
						Email string `gorm:"size:100;uniqueIndex"`
					}{},
					Previous: nil,
				},
			},
			wantSQLs: map[string][]string{
				"1.0.0_up.sql": {
					"CREATE TABLE",
					`"name"`,
					`"email"`,
					"CREATE UNIQUE INDEX",
				},
				"1.0.0_down.sql": {
					"DROP TABLE",
				},
			},
		},
		{
			name: "add column and index",
			versions: []models.ModelVersion{
				{
					Version: "1.0.0",
					Current: &struct {
						TestModel
						Name string `gorm:"size:100"`
					}{},
					Previous: nil,
				},
				{
					Version: "1.0.1",
					Current: &struct {
						TestModel
						Name  string `gorm:"size:100;index:idx_name"`
						Email string `gorm:"size:100"` // new column
					}{},
					Previous: &struct {
						TestModel
						Name string `gorm:"size:100"`
					}{},
				},
			},
			wantSQLs: map[string][]string{
				"1.0.0_up.sql": {
					"CREATE TABLE",
					`"name"`,
				},
				"1.0.0_down.sql": {
					"DROP TABLE",
				},
				"1.0.1_up.sql": {
					"ALTER TABLE",
					`ADD "email"`,
					"CREATE INDEX",
					`"idx_name"`,
				},
				"1.0.1_down.sql": {
					"DROP INDEX",
					"DROP COLUMN",
					`"email"`,
				},
			},
		},
		{
			name: "composite index",
			versions: []models.ModelVersion{
				{
					Version: "1.0.0",
					Current: &struct {
						TestModel
						Name  string `gorm:"size:100"`
						Email string `gorm:"size:100"`
					}{},
					Previous: nil,
				},
				{
					Version: "1.0.1",
					Current: &struct {
						TestModel
						Name  string `gorm:"size:100;index:idx_name_email,priority:1"`
						Email string `gorm:"size:100;index:idx_name_email,priority:2"`
					}{},
					Previous: &struct {
						TestModel
						Name  string `gorm:"size:100"`
						Email string `gorm:"size:100"`
					}{},
				},
			},
			wantSQLs: map[string][]string{
				"1.0.0_up.sql": {
					"CREATE TABLE",
					`"name"`,
					`"email"`,
				},
				"1.0.0_down.sql": {
					"DROP TABLE",
				},
				"1.0.1_up.sql": {
					"CREATE INDEX",
					`"idx_name_email"`,
					`"name"`,
					`"email"`,
				},
				"1.0.1_down.sql": {
					"DROP INDEX",
					`"idx_name_email"`,
				},
			},
		},
		{
			name: "deprecate columns",
			skip: true,
			versions: []models.ModelVersion{
				{
					Version: "1.0.0",
					Current: &struct {
						TestModel
						Name    string `gorm:"size:100"`
						Email   string `gorm:"size:100"`
						Address string `gorm:"size:200"`
						Phone   string `gorm:"size:20"`
					}{},
					Previous: nil,
				},
				{
					Version: "1.0.1",
					Current: &struct {
						TestModel
						Name  string `gorm:"size:100"`
						Email string `gorm:"size:100"`
					}{},
					Previous: &struct {
						TestModel
						Name    string `gorm:"size:100"`
						Email   string `gorm:"size:100"`
						Address string `gorm:"size:200"`
						Phone   string `gorm:"size:20"`
					}{},
				},
			},
			wantSQLs: map[string][]string{
				"1.0.0_up.sql": {
					"CREATE TABLE",
					`"test_models"`,
					`"name" character varying(100)`,
					`"email" character varying(100)`,
					`"address" character varying(200)`,
					`"phone" character varying(20)`,
				},
				"1.0.0_down.sql": {
					"DROP TABLE IF EXISTS",
					`"test_models"`,
				},
				"1.0.1_up.sql": {
					"ALTER TABLE",
					`"test_models"`,
					`DROP COLUMN IF EXISTS "address"`,
					`DROP COLUMN IF EXISTS "phone"`,
				},
				"1.0.1_down.sql": {
					"ALTER TABLE",
					`"test_models"`,
					`ADD COLUMN "address" character varying(200)`,
					`ADD COLUMN "phone" character varying(20)`,
				},
			},
		},
		{
			name:     "empty migration",
			versions: []models.ModelVersion{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		if tt.skip {
			t.Skipf("Skipping test '%s'", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			sqls, err := GenerateSQL(db, tt.versions)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.wantSQLs), len(sqls), "Number of generated SQL files does not match\nWant: %v\nGot: %v", tt.wantSQLs, sqls)

			// Verify SQL content for each file
			for filename, wantPhrases := range tt.wantSQLs {
				sql, ok := sqls[filename]
				assert.True(t, ok, "Missing SQL file '%s'\nWant: %v\nGot: %v", filename, tt.wantSQLs, sqls)

				// Verify required SQL statements
				for _, phrase := range wantPhrases {
					assert.True(t, strings.Contains(strings.ToUpper(sql), strings.ToUpper(phrase)),
						"SQL should contain '%s'\nWant SQL phrases: %v\nGot SQL: %s", phrase, wantPhrases, sql)
				}
			}
		})
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
