package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	// Hash the password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Create the user
	result := r.db.WithContext(ctx).Create(user)
	return result.Error
}

// Update updates a user in the database
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	// If password is being updated, hash it
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	result := r.db.WithContext(ctx).Save(user)
	return result.Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, result.Error
	}
	return &user, nil
}

// List retrieves a list of users with pagination
func (r *UserRepository) List(ctx context.Context, skip, limit int) ([]*models.User, error) {
	var users []*models.User
	result := r.db.WithContext(ctx).Offset(skip).Limit(limit).Find(&users)
	return users, result.Error
}

// Delete deletes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	return result.Error
}

// CleanupTestData removes all test data from the database
func (r *UserRepository) CleanupTestData(ctx context.Context) error {
	// Delete all users, including soft-deleted ones
	result := r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}

	// Reset the auto-increment sequence if any
	err := r.db.WithContext(ctx).Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
	return err
}

// Remove the old helper functions
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Helper function to check if password matches hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
