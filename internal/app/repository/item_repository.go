package repository

import (
	"context"
	"errors"

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

func (r *ItemRepository) Update(ctx context.Context, item *models.Item) error {
	result := r.db.WithContext(ctx).Save(item)
	return result.Error
}

func (r *ItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Item{}, "id = ?", id)
	return result.Error
}

func (r *ItemRepository) List(ctx context.Context, skip, limit int) ([]*models.Item, error) {
	var items []*models.Item
	result := r.db.WithContext(ctx).Offset(skip).Limit(limit).Find(&items)
	return items, result.Error
}

func (r *ItemRepository) CleanupTestData(ctx context.Context) error {
	result := r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.Item{})
	if result.Error != nil {
		return result.Error
	}
	return r.db.WithContext(ctx).Exec("TRUNCATE TABLE items RESTART IDENTITY CASCADE").Error
}
