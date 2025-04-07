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

// table name
func (TestModel) TableName() string {
	return "test_models"
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
			wantErr: true, // Mark this test case as expected to fail
		},
		{
			name:     "empty migration",
			versions: []models.ModelVersion{},
			wantErr:  true,
		},
		{
			name: "add foreign key",
			skip: true,
			versions: []models.ModelVersion{
				{
					Version: "1.0.0",
					Current: &struct {
						TestModel
						CompanyID uint   `gorm:"column:company_id"`
						Name      string `gorm:"size:100"`
						Company   struct {
							ID   uint   `gorm:"primarykey"`
							Name string `gorm:"size:100"`
						} `gorm:"foreignKey:CompanyID"`
					}{},
					Previous: nil,
				},
				{
					Version: "1.0.1",
					Current: &struct {
						TestModel
						CompanyID uint   `gorm:"column:company_id;index:idx_company"`
						Name      string `gorm:"size:100"`
						Company   struct {
							ID   uint   `gorm:"primarykey"`
							Name string `gorm:"size:100"`
						} `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
					}{},
					Previous: &struct {
						TestModel
						CompanyID uint   `gorm:"column:company_id"`
						Name      string `gorm:"size:100"`
						Company   struct {
							ID   uint   `gorm:"primarykey"`
							Name string `gorm:"size:100"`
						} `gorm:"foreignKey:CompanyID"`
					}{},
				},
			},
			wantSQLs: map[string][]string{
				"1.0.0_up.sql": {
					"CREATE TABLE",
					`"company_id"`,
					`"name"`,
					"FOREIGN KEY",
					"REFERENCES",
				},
				"1.0.0_down.sql": {
					"DROP TABLE",
				},
				"1.0.1_up.sql": {
					"ALTER TABLE",
					"CREATE INDEX",
					`"idx_company"`,
					"ON DELETE CASCADE",
				},
				"1.0.1_down.sql": {
					"DROP INDEX",
					"ALTER TABLE",
				},
			},
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

			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, len(tt.wantSQLs), len(sqls), "Number of generated SQL files does not match\nWant: %v\nGot: %v", tt.wantSQLs, sqls) {
				return
			}

			// Verify SQL content for each file
			for filename, wantPhrases := range tt.wantSQLs {
				sql, ok := sqls[filename]
				if !assert.True(t, ok, "Missing SQL file '%s'\nWant: %v\nGot: %v", filename, tt.wantSQLs, sqls) {
					return
				}

				// Verify required SQL statements
				for _, phrase := range wantPhrases {
					if !assert.True(t, strings.Contains(strings.ToUpper(sql), strings.ToUpper(phrase)),
						"SQL should contain '%s'\nWant SQL phrases: %v\nGot SQL: %s", phrase, wantPhrases, sql) {
						return
					}
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
