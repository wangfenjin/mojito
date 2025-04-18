// Package migrations provides functionality for database schema migrations
package migrations

import (
	"fmt"
	"sort"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GenerateMigration generates a slice of gormigrate.Migration based on the provided models
func GenerateMigration(models []ModelVersion) []*gormigrate.Migration {
	var migrationsList []*gormigrate.Migration

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
					return tx.Migrator().CreateTable(mv.Current)
				}
				return handleChanges(tx, mv.Previous, mv.Current)
			},
			Rollback: func(tx *gorm.DB) error {
				if mv.Previous == nil {
					return tx.Migrator().DropTable(mv.Current)
				}
				return handleChanges(tx, mv.Current, mv.Previous)
			},
		}
		migrationsList = append(migrationsList, migration)
	}
	return migrationsList
}

func handleChanges(tx *gorm.DB, oldModel, newModel interface{}) error {
	if err := handleColumnChanges(tx, oldModel, newModel); err != nil {
		return err
	}

	if err := handleConstraints(tx, oldModel, newModel); err != nil {
		return err
	}

	if err := handleIndexChanges(tx, oldModel, newModel); err != nil {
		return err
	}

	return nil
}

func handleColumnChanges(tx *gorm.DB, oldModel, newModel interface{}) error {
	oldStmt := &gorm.Statement{DB: tx}
	if err := oldStmt.Parse(oldModel); err != nil {
		return fmt.Errorf("parsing old model: %w", err)
	}

	newStmt := &gorm.Statement{DB: tx}
	if err := newStmt.Parse(newModel); err != nil {
		return fmt.Errorf("parsing new model: %w", err)
	}

	sort.Slice(oldStmt.Schema.Fields, func(i, j int) bool {
		return oldStmt.Schema.Fields[i].Name < oldStmt.Schema.Fields[j].Name
	})
	sort.Slice(newStmt.Schema.Fields, func(i, j int) bool {
		return newStmt.Schema.Fields[i].Name < newStmt.Schema.Fields[j].Name
	})

	// Handle dropped columns first
	for _, oldField := range oldStmt.Schema.Fields {
		if newStmt.Schema.LookUpField(oldField.Name) == nil {
			if err := tx.Migrator().DropColumn(newModel, oldField.DBName); err != nil {
				return fmt.Errorf("dropping column %s: %w", oldField.DBName, err)
			}
		}
	}

	// Handle new columns and modifications
	for _, newField := range newStmt.Schema.Fields {
		oldField := oldStmt.Schema.LookUpField(newField.Name)

		if oldField == nil {
			if err := tx.Migrator().AddColumn(newModel, newField.DBName); err != nil {
				return fmt.Errorf("adding column %s: %w", newField.DBName, err)
			}
		} else if columnNeedsAlter(oldField, newField) {
			if err := tx.Migrator().AlterColumn(newModel, newField.DBName); err != nil {
				return fmt.Errorf("altering column %s: %w", newField.DBName, err)
			}
		}
	}
	return nil
}

func handleConstraints(tx *gorm.DB, oldModel, newModel interface{}) error {
	oldStmt := &gorm.Statement{DB: tx}
	newStmt := &gorm.Statement{DB: tx}

	if err := oldStmt.Parse(oldModel); err != nil {
		return err
	}
	if err := newStmt.Parse(newModel); err != nil {
		return err
	}

	sort.Slice(oldStmt.Schema.Fields, func(i, j int) bool {
		return oldStmt.Schema.Fields[i].Name < oldStmt.Schema.Fields[j].Name
	})
	sort.Slice(newStmt.Schema.Fields, func(i, j int) bool {
		return newStmt.Schema.Fields[i].Name < newStmt.Schema.Fields[j].Name
	})

	// Drop old foreign key constraints
	for _, field := range oldStmt.Schema.Fields {
		if _, ok := field.TagSettings["FOREIGNKEY"]; ok {
			constraintName := fmt.Sprintf("fk_%s_%s", oldStmt.Schema.Table, field.DBName)
			newField := newStmt.Schema.LookUpField(field.Name)
			if newField == nil {
				if tx.Migrator().HasConstraint(oldModel, constraintName) {
					if err := tx.Migrator().DropConstraint(oldModel, constraintName); err != nil {
						return err
					}
				}
			}
		}
	}

	// create new foreign key constraints
	for _, field := range newStmt.Schema.Fields {
		constraintName := fmt.Sprintf("uk_%s_%s", newStmt.Schema.Table, field.DBName)
		if !tx.Migrator().HasConstraint(newModel, constraintName) {
			if err := tx.Migrator().CreateConstraint(newModel, constraintName); err != nil {
				return fmt.Errorf("creating unique constraint %s: %w", constraintName, err)
			}
		}
	}

	return nil
}

func columnNeedsAlter(oldField, newField *schema.Field) bool {
	// Compare data type
	if oldField.DataType != newField.DataType {
		return true
	}

	// Compare size for variable length types
	if oldField.Size != newField.Size {
		return true
	}

	// Compare nullability
	if oldField.NotNull != newField.NotNull {
		return true
	}

	// Compare default value
	if oldField.HasDefaultValue != newField.HasDefaultValue {
		return true
	}

	// Compare auto timestamp settings
	oldAutoCreate := oldField.AutoCreateTime > 0
	newAutoCreate := newField.AutoCreateTime > 0
	if oldAutoCreate != newAutoCreate {
		return true
	}

	oldAutoUpdate := oldField.AutoUpdateTime > 0
	newAutoUpdate := newField.AutoUpdateTime > 0
	return oldAutoUpdate != newAutoUpdate
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

	// TODO: rename indexes
	// for _, oldIdx := range oldIndexes {
	// 	if !hasIndex(newIndexes, oldIdx.Name) {
	// 		for _, newIdx := range newIndexes {
	// 			if !hasIndex(oldIndexes, newIdx.Name) &&
	// 				len(oldIdx.Fields) == len(newIdx.Fields) &&
	// 				oldIdx.Type == newIdx.Type {
	// 				fieldsMatch := true
	// 				for i, oldField := range oldIdx.Fields {
	// 					if oldField != newIdx.Fields[i] {
	// 						fieldsMatch = false
	// 						break
	// 					}
	// 				}

	// 				if fieldsMatch {
	// 					if err := tx.Migrator().RenameIndex(newModel, oldIdx.Name, newIdx.Name); err != nil {
	// 						return fmt.Errorf("renaming index %s to %s: %w", oldIdx.Name, newIdx.Name, err)
	// 					} else {
	// 						// TODO: remove renamed index from newIndexes and oldIndexes and continue with next index in oldIndexes
	// 						break
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

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

func hasIndex(indexes []*schema.Index, name string) bool {
	for _, idx := range indexes {
		if idx.Name == name {
			return true
		}
	}
	return false
}
