package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userBase struct{}

func (userBase) TableName() string {
	return "users"
}

// User represents a user in the system
type UserV1 struct {
	userBase
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	Email       string         `gorm:"uniqueIndex;type:varchar(255);not null"`
	Password    string         `gorm:"type:varchar(255);not null"`
	FullName    string         `gorm:"type:varchar(100);not null"`
	IsActive    bool           `gorm:"default:true"`
	IsSuperuser bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_users_created"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_users_updated"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// User represents a user in the system
type UserV2 struct {
	userBase
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	Email       string         `gorm:"uniqueIndex:idx_users_email;type:varchar(255);not null" validate:"required,email,max=255"`
	PhoneNumber string         `gorm:"uniqueIndex:idx_users_phone;type:varchar(20)" validate:"omitempty,e164"`
	Password    string         `gorm:"type:varchar(255);not null" validate:"required,min=8"`
	FullName    string         `gorm:"type:varchar(100);not null" validate:"required,min=2,max=100"`
	IsActive    bool           `gorm:"default:true"`
	IsSuperuser bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_users_created;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index:idx_users_updated;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type User = UserV2

// BeforeCreate will set a UUID rather than numeric ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
