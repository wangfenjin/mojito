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

// UserV1 represents the first version of the user model
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

// UserV2 represents the second version of the user model with additional fields
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

// User is the model for the user table
// Update this to the latest version of the user dat
type User = UserV2

// BeforeCreate will set a UUID rather than numeric ID
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
