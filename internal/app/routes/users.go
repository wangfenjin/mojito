package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/repository"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

// RegisterUsersRoutes registers all user related routes
func RegisterUsersRoutes(r *gin.Engine) {
	usersGroup := r.Group("/api/v1/users", middleware.RequireAuth())
	{
		// Protected routes (require auth)
		usersGroup.GET("/",
			middleware.WithHandler(listUsersHandler))

		usersGroup.GET("/me",
			middleware.WithHandler(getCurrentUserHandler))

		usersGroup.DELETE("/me",
			middleware.WithHandler(deleteCurrentUserHandler))

		usersGroup.PATCH("/me",
			middleware.WithHandler(updateCurrentUserHandler))

		usersGroup.PATCH("/me/password",
			middleware.WithHandler(updatePasswordHandler))

		usersGroup.GET("/:id",
			middleware.WithHandler(getUserHandler))

		usersGroup.PATCH("/:id",
			middleware.WithHandler(updateUserHandler))
	}

	// Public routes (no auth required)
	r.POST("/api/v1/users/signup",
		middleware.WithHandler(registerUserHandler))
}

// CreateUserRequest
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

// UpdateUserRequest
type UpdateUserRequest struct {
	ID          string `uri:"id" binding:"required,uuid"`
	Email       string `json:"email" binding:"omitempty,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

// RegisterUserRequest
type RegisterUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
}

// Update the original UpdateUserMeRequest
type UpdateUserMeRequest struct {
	Email       string `json:"email" binding:"omitempty,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
	FullName    string `json:"full_name"`
}

type GetUserRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type ListUsersRequest struct {
	Skip  int `form:"skip" binding:"min=0" default:"0"`
	Limit int `form:"limit" binding:"min=1,max=100" default:"10"`
}

// UserResponse represents the standard user response format
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	FullName    string    `json:"full_name"`
	IsActive    bool      `json:"is_active"`
	IsSuperuser bool      `json:"is_superuser"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UsersResponse represents a paginated list of users
type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Meta  struct {
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// Add new handlers
func deleteCurrentUserHandler(ctx context.Context, _ any) (*MessageResponse, error) {
	// Get current user ID from context
	userID := ctx.Value("user_id").(string)
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	err = userRepo.Delete(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error deleting user: %w", err)
	}
	return &MessageResponse{Message: "User deleted successfully"}, nil
}

func updatePasswordHandler(ctx context.Context, req UpdatePasswordRequest) (*MessageResponse, error) {
	userID := ctx.Value("user_id").(string)
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	// Get user with current password hash from DB
	user, err := userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Verify current password against stored hash
	if !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
		return nil, middleware.NewBadRequestError("incorrect current password")
	}

	// Hash the new password
	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Update with new password hash
	user.Password = hashedNewPassword
	err = userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error updating password: %w", err)
	}

	return &MessageResponse{
		Message: "Password updated successfully",
	}, nil
}

// Update handler functions
func registerUserHandler(ctx context.Context, req RegisterUserRequest) (*UserResponse, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Check if user with this email already exists
	existingUser, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		return nil, middleware.NewBadRequestError("user with this email already exists")
	}

	// Check phone number if provided
	if req.PhoneNumber != "" {
		existingUser, err = userRepo.GetByPhone(ctx, req.PhoneNumber)
		if err != nil {
			return nil, fmt.Errorf("error checking existing phone: %w", err)
		}
		if existingUser != nil {
			return nil, middleware.NewBadRequestError("user with this phone number already exists")
		}
	}

	user := &models.User{
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		FullName:    req.FullName,
		IsActive:    true,
		IsSuperuser: false,
	}

	err = userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Update response maps in other handlers to include phone_number
func updateCurrentUserHandler(ctx context.Context, req UpdateUserMeRequest) (*UserResponse, error) {
	userID := ctx.Value("user_id").(string)
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	user, err := userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.PhoneNumber != "" {
		// Check if phone number is already used by another user
		existingUser, err := userRepo.GetByPhone(ctx, req.PhoneNumber)
		if err != nil {
			return nil, fmt.Errorf("error checking existing phone: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, middleware.NewBadRequestError("phone number already in use")
		}
		user.PhoneNumber = req.PhoneNumber
	}

	err = userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Update getCurrentUserHandler response
func getCurrentUserHandler(ctx context.Context, _ any) (*UserResponse, error) {
	userID := ctx.Value("user_id").(string)
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	user, err := userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, middleware.NewBadRequestError("user not found")
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Update getUserHandler response
func getUserHandler(ctx context.Context, req GetUserRequest) (*UserResponse, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID format")
	}

	user, err := userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, middleware.NewBadRequestError("user not found")
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Update updateUserHandler to handle phone number
func updateUserHandler(ctx context.Context, req UpdateUserRequest) (*UserResponse, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// check if user is superuser
	userID := ctx.Value("user_id").(string)
	currentID, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID format")
	}
	currentUser, err := userRepo.GetByID(ctx, currentID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if !currentUser.IsSuperuser {
		return nil, middleware.NewForbiddenError("only superusers can update other users")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID format")
	}

	// Get existing user
	user, err := userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, middleware.NewBadRequestError("user not found")
	}

	// Explicitly reject email updates
	if req.Email != "" {
		return nil, middleware.NewBadRequestError("email updates are not allowed")
	}

	// Update fields if provided
	if req.Password != "" {
		// Hash the password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("error hashing password: %w", err)
		}
		user.Password = hashedPassword
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	user.IsActive = req.IsActive
	user.IsSuperuser = req.IsSuperuser

	// Save updates
	err = userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FullName:    user.FullName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Update listUsersHandler response
func listUsersHandler(ctx context.Context, req ListUsersRequest) (*UsersResponse, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Set default values for pagination
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Skip < 0 {
		req.Skip = 0
	}

	users, err := userRepo.List(ctx, req.Skip, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Convert to response format
	userList := make([]UserResponse, len(users))
	for i, user := range users {
		userList[i] = UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			FullName:    user.FullName,
			IsActive:    user.IsActive,
			IsSuperuser: user.IsSuperuser,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
	}

	return &UsersResponse{
		Users: userList,
		Meta: struct {
			Skip  int `json:"skip"`
			Limit int `json:"limit"`
		}{
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	}, nil
}
