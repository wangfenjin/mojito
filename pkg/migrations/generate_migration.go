package migrations

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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

					if err := handleConstraints(tx, mv.Previous, mv.Current); err != nil {
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
		migrationsList = append(migrationsList, migration)
	}
	return migrationsList
}

// Remove handleTableRename function

func handleColumnChanges(tx *gorm.DB, model interface{}, oldType, newType reflect.Type) error {
	oldStmt := &gorm.Statement{DB: tx}
	if err := oldStmt.Parse(reflect.New(oldType).Interface()); err != nil {
		return fmt.Errorf("parsing old model: %w", err)
	}

	newStmt := &gorm.Statement{DB: tx}
	if err := newStmt.Parse(model); err != nil {
		return fmt.Errorf("parsing new model: %w", err)
	}

	// Handle dropped columns first
	for _, oldField := range oldStmt.Schema.Fields {
		if newStmt.Schema.LookUpField(oldField.Name) == nil {
			if err := tx.Migrator().DropColumn(model, oldField.DBName); err != nil {
				return fmt.Errorf("dropping column %s: %w", oldField.DBName, err)
			}
		}
	}

	// Handle new columns and modifications
	for _, newField := range newStmt.Schema.Fields {
		oldField := oldStmt.Schema.LookUpField(newField.Name)

		if oldField == nil {
			if err := tx.Migrator().AddColumn(model, newField.DBName); err != nil {
				return fmt.Errorf("adding column %s: %w", newField.DBName, err)
			}
		} else if columnNeedsAlter(oldField, newField) {
			if err := tx.Migrator().AlterColumn(model, newField.DBName); err != nil {
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

	// Handle foreign key constraints
	for _, field := range newStmt.Schema.Fields {
		if field.TagSettings["FOREIGNKEY"] != "" {
			// Get reference table and field
			refSchema := &gorm.Statement{DB: tx}
			if err := refSchema.Parse(reflect.New(field.IndirectFieldType).Interface()); err != nil {
				return fmt.Errorf("parsing reference model: %w", err)
			}

			constraintName := fmt.Sprintf("fk_%s_%s", newStmt.Schema.Table, field.DBName)

			// Drop existing constraint if exists
			if tx.Migrator().HasConstraint(newModel, constraintName) {
				if err := tx.Migrator().DropConstraint(newModel, constraintName); err != nil {
					log.Printf("Warning: dropping foreign key constraint %s: %v", constraintName, err)
				}
			}

			// Use raw SQL only if we have special options
			if constraint := field.TagSettings["CONSTRAINT"]; constraint != "" &&
				(strings.Contains(constraint, "OnDelete:") || strings.Contains(constraint, "OnUpdate:")) {
				sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(id)",
					newStmt.Schema.Table, constraintName, field.DBName, refSchema.Schema.Table)
				if strings.Contains(constraint, "OnDelete:CASCADE") {
					sql += " ON DELETE CASCADE"
				}
				if strings.Contains(constraint, "OnUpdate:CASCADE") {
					sql += " ON UPDATE CASCADE"
				}
				if err := tx.Exec(sql).Error; err != nil {
					return fmt.Errorf("creating foreign key constraint with options: %w", err)
				}
			} else {
				// Use GORM's API for basic foreign key constraints
				if err := tx.Migrator().CreateConstraint(newModel, constraintName); err != nil {
					return fmt.Errorf("creating foreign key constraint: %w", err)
				}
			}
		}
	}

	// Drop old foreign key constraints
	for _, field := range oldStmt.Schema.Fields {
		if _, ok := field.TagSettings["FOREIGNKEY"]; ok {
			constraintName := fmt.Sprintf("fk_%s_%s", oldStmt.Schema.Table, field.DBName)
			newField := newStmt.Schema.LookUpField(field.Name)
			if newField == nil {
				if tx.Migrator().HasConstraint(oldModel, constraintName) {
					if err := tx.Migrator().DropConstraint(oldModel, constraintName); err != nil {
						log.Printf("Warning: dropping old foreign key constraint %s: %v", constraintName, err)
						return err
					}
				}
			}
		}
	}

	// Handle unique constraints
	for _, field := range newStmt.Schema.Fields {
		if field.Unique {
			constraintName := fmt.Sprintf("uk_%s_%s", newStmt.Schema.Table, field.DBName)
			if !tx.Migrator().HasConstraint(newModel, constraintName) {
				if err := tx.Migrator().CreateConstraint(newModel, constraintName); err != nil {
					log.Printf("Warning: creating unique constraint %s: %v", constraintName, err)
				}
			}
		}
	}

	// Check for removed unique constraints
	for _, field := range oldStmt.Schema.Fields {
		if field.Unique {
			constraintName := fmt.Sprintf("uk_%s_%s", oldStmt.Schema.Table, field.DBName)
			newField, exists := newStmt.Schema.FieldsByDBName[field.DBName]
			if !exists || !newField.Unique {
				if tx.Migrator().HasConstraint(oldModel, constraintName) {
					if err := tx.Migrator().DropConstraint(oldModel, constraintName); err != nil {
						log.Printf("Warning: dropping unique constraint %s: %v", constraintName, err)
					}
				}
			}
		}
	}

	return nil
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

	// 创建约束
	for _, field := range stmt.Schema.Fields {
		if field.Unique {
			constraintName := fmt.Sprintf("uk_%s_%s", stmt.Schema.Table, field.DBName)
			if err := tx.Migrator().CreateConstraint(model, constraintName); err != nil {
				log.Printf("Warning: creating constraint %s: %v", constraintName, err)
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

// 辅助函数：从标签中获取值
func getTagValue(tag reflect.StructTag, tagName, key string) (string, bool) {
	gormTag := tag.Get(tagName)
	if gormTag == "" {
		return "", false
	}

	for _, option := range strings.Split(gormTag, ";") {
		option = strings.TrimSpace(option)
		if strings.HasPrefix(option, key+":") {
			return strings.TrimPrefix(option, key+":"), true
		}
	}
	return "", false
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

	// 检查索引重命名
	for oldName, oldIdx := range oldIndexes {
		if !hasIndex(newIndexes, oldName) {
			// 查找可能是重命名的索引
			for newName, newIdx := range newIndexes {
				if !hasIndex(oldIndexes, newName) &&
					len(oldIdx.Fields) == len(newIdx.Fields) &&
					oldIdx.Type == newIdx.Type {
					// 可能是重命名的索引，检查字段是否匹配
					fieldsMatch := true
					for i, oldField := range oldIdx.Fields {
						if oldField != newIdx.Fields[i] {
							fieldsMatch = false
							break
						}
					}

					if fieldsMatch {
						if err := tx.Migrator().RenameIndex(newModel, oldName, newName); err != nil {
							log.Printf("Warning: renaming index %s to %s: %v", oldName, newName, err)
						} else {
							log.Printf("索引已重命名: %s -> %s", oldName, newName)
							// 从旧索引列表中移除，避免后续被删除
							delete(oldIndexes, oldName)
							break
						}
					}
				}
			}
		}
	}

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

	// 处理约束回滚
	if err := handleConstraints(tx, newModel, oldModel); err != nil {
		return err
	}

	// Drop new columns
	for i := 0; i < newType.NumField(); i++ {
		newField := newType.Field(i)
		if _, exists := oldType.FieldByName(newField.Name); !exists {
			// 检查是否是重命名的列
			isRenamed := false
			if renameFrom, ok := getTagValue(newField.Tag, "gorm", "rename"); ok && renameFrom != "" {
				// 对于重命名的列，需要恢复原名
				if err := tx.Migrator().RenameColumn(newModel, newField.Name, renameFrom); err != nil {
					return err
				}
				isRenamed = true
			}

			if !isRenamed {
				if err := tx.Migrator().DropColumn(newModel, newField.Name); err != nil {
					return err
				}
			}
		}
	}

	// Restore old indexes
	return handleIndexChanges(tx, newModel, oldModel)
}

func hasIndex(indexes map[string]schema.Index, name string) bool {
	if _, ok := indexes[name]; ok {
		return ok
	}
	return false
}

// Remove GenerateMigrationSQL function as it's no longer used
