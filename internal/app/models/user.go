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
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Email       string    `gorm:"uniqueIndex;not null"`
	Password    string    `gorm:"not null"`
	FullName    string    `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	IsSuperuser bool      `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// User represents a user in the system
type UserV2 struct {
	userBase
	ID          uuid.UUID `gorm:"type:uuid;primary_key;not null"`
	Email       string    `gorm:"uniqueIndex:idx_users_email;type:varchar(255);not null" validate:"required,email"`
	PhoneNumber string    `gorm:"uniqueIndex:idx_users_phone;type:varchar(20)" validate:"omitempty,e164"`
	Password    string    `gorm:"type:varchar(255);not null"`
	FullName    string    `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	IsSuperuser bool      `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
