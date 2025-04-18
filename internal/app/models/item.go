package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type itemBase struct{}

func (itemBase) TableName() string {
	return "items"
}

// ItemV1 is the model for the item table
type ItemV1 struct {
	itemBase
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	Title       string         `gorm:"type:varchar(200);not null"`
	Description string         `gorm:"type:text"`
	OwnerID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_items_owner;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" validate:"required"`
	Owner       User           `gorm:"foreignKey:OwnerID;references:ID"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_items_created;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_items_updated;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Item is the model for the item table
type Item = ItemV1

// BeforeCreate will set a UUID rather than numeric ID
func (i *Item) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
