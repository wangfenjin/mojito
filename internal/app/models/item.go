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

type ItemV1 struct {
	itemBase
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `gorm:"not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Item = ItemV1
