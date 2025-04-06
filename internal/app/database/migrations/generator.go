package migrations

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func GenerateMigration(models []models.ModelVersion) []*gormigrate.Migration {
	var migrations []*gormigrate.Migration

	// No need for initial migration to create version table anymore

	// Sort models by version to ensure correct order
	sort.Slice(models, func(i, j int) bool {
		return models[i].Version < models[j].Version
	})

	for _, mv := range models {
		migration := &gormigrate.Migration{
			ID: mv.Version,
			Migrate: func(tx *gorm.DB) error {
				if mv.Previous == nil {
					if err := tx.Migrator().CreateTable(mv.Current); err != nil {
						log.Printf("Warning: creating table: %v", err)
					}
					if err := createIndexesAndConstraints(tx, mv.Current); err != nil {
						log.Printf("Warning: creating indexes: %v", err)
					}
				} else {
					if err := handleColumnChanges(tx, mv.Current, reflect.TypeOf(mv.Previous).Elem(), reflect.TypeOf(mv.Current).Elem()); err != nil {
						return err
					}
					if err := handleIndexChanges(tx, mv.Previous, mv.Current); err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// TODO: how to capture the rollback SQL?
				if mv.Previous == nil {
					return tx.Migrator().DropTable(mv.Current)
				}
				return handleRollback(tx, mv.Previous, mv.Current)
			},
		}
		migrations = append(migrations, migration)
	}
	return migrations
}

func createIndexesAndConstraints(tx *gorm.DB, model interface{}) error {
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(model); err != nil {
		return fmt.Errorf("parsing model: %w", err)
	}

	// Create indexes
	for _, idx := range stmt.Schema.ParseIndexes() {
		if err := tx.Migrator().CreateIndex(model, idx.Name); err != nil {
			log.Printf("Warning: creating index %s: %v", idx.Name, err)
		}
	}

	return nil
}

func handleColumnChanges(tx *gorm.DB, model interface{}, oldType, newType reflect.Type) error {
	for i := 0; i < newType.NumField(); i++ {
		newField := newType.Field(i)
		oldField, exists := oldType.FieldByName(newField.Name)

		if !exists {
			// New column
			if err := tx.Migrator().AddColumn(model, newField.Name); err != nil {
				return err
			}
		} else {
			// Check for type changes or constraints
			if needsAlter(oldField, newField) {
				if err := tx.Migrator().AlterColumn(model, newField.Name); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleIndexChanges(tx *gorm.DB, oldModel, newModel interface{}) error {
	oldStmt := &gorm.Statement{DB: tx}
	newStmt := &gorm.Statement{DB: tx}

	if err := oldStmt.Parse(oldModel); err != nil {
		return err
	}
	if err := newStmt.Parse(newModel); err != nil {
		return err
	}

	// Compare and update indexes
	oldIndexes := oldStmt.Schema.ParseIndexes()
	newIndexes := newStmt.Schema.ParseIndexes()

	// Drop removed indexes
	for _, oldIdx := range oldIndexes {
		if !hasIndex(newIndexes, oldIdx.Name) {
			if err := tx.Migrator().DropIndex(newModel, oldIdx.Name); err != nil {
				return err
			}
		}
	}

	// Create new indexes
	for _, newIdx := range newIndexes {
		if !hasIndex(oldIndexes, newIdx.Name) {
			if err := tx.Migrator().CreateIndex(newModel, newIdx.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func handleRollback(tx *gorm.DB, oldModel, newModel interface{}) error {
	oldType := reflect.TypeOf(oldModel).Elem()
	newType := reflect.TypeOf(newModel).Elem()

	// Drop new columns
	for i := 0; i < newType.NumField(); i++ {
		newField := newType.Field(i)
		if _, exists := oldType.FieldByName(newField.Name); !exists {
			if err := tx.Migrator().DropColumn(newModel, newField.Name); err != nil {
				return err
			}
		}
	}

	// Restore old indexes
	return handleIndexChanges(tx, newModel, oldModel)
}

func needsAlter(oldField, newField reflect.StructField) bool {
	oldTag := oldField.Tag.Get("gorm")
	newTag := newField.Tag.Get("gorm")
	return oldTag != newTag || oldField.Type != newField.Type
}

func hasIndex(indexes map[string]schema.Index, name string) bool {
	if _, ok := indexes[name]; ok {
		return ok
	}
	return false
}

// Remove GenerateMigrationSQL function as it's no longer used
