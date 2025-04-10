package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/models"
	"gorm.io/gorm"
)

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	if item.OwnerID == uuid.Nil {
		return fmt.Errorf("owner ID is required")
	}
	result := r.db.WithContext(ctx).Create(item)
	return result.Error
}

func (r *ItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	var item models.Item
	result := r.db.WithContext(ctx).First(&item, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *ItemRepository) GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*models.Item, error) {
	var item models.Item
	result := r.db.WithContext(ctx).First(&item, "id = ? AND owner_id = ?", id, ownerID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *ItemRepository) Update(ctx context.Context, item *models.Item) error {
	if item.OwnerID == uuid.Nil {
		return fmt.Errorf("owner ID is required")
	}
	result := r.db.WithContext(ctx).Save(item)
	return result.Error
}

func (r *ItemRepository) Delete(ctx context.Context, id, ownerID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Item{}, "id = ? AND owner_id = ?", id, ownerID)
	if result.RowsAffected == 0 {
		return fmt.Errorf("item not found or not owned by user")
	}
	return result.Error
}

func (r *ItemRepository) List(ctx context.Context, ownerID uuid.UUID, skip, limit int) ([]*models.Item, error) {
	var items []*models.Item
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Offset(skip).
		Limit(limit).
		Find(&items)
	return items, result.Error
}

func (r *ItemRepository) CleanupTestData(ctx context.Context) error {
	// First delete all records with global update allowed
	result := r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Item{})
	return result.Error
}
