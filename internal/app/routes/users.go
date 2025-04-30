package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/models/gen"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

// RegisterUsersRoutes registers all user related routes
func RegisterUsersRoutes(r chi.Router) {
	// Protected routes (require auth)
	r.Route("/api/v1/users", func(r chi.Router) {
		// Apply auth middleware to all routes in this group
		r.Use(middleware.RequireAuth())

		r.Get("/", middleware.WithHandler(listUsersHandler))
		r.Get("/me", middleware.WithHandler(getCurrentUserHandler))
		r.Delete("/me", middleware.WithHandler(deleteCurrentUserHandler))
		r.Patch("/me", middleware.WithHandler(updateCurrentUserHandler))
		r.Patch("/me/password", middleware.WithHandler(updatePasswordHandler))
		r.Get("/{id}", middleware.WithHandler(getUserHandler))
		r.Patch("/{id}", middleware.WithHandler(updateUserHandler))
	})

	// Public routes (no auth required)
	r.Post("/api/v1/users/signup", middleware.WithHandler(registerUserHandler))
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	ID          string `uri:"id" binding:"required,uuid"`
	Email       string `json:"email" binding:"omitempty,email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

// RegisterUserRequest represents the request body for user registration
type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

// UpdateUserMeRequest represents the request body for updating the current user
type UpdateUserMeRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
}

// GetUserRequest represents the request parameters for getting a user
type GetUserRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// ListUsersRequest represents the request parameters for listing users
type ListUsersRequest struct {
	Skip  int64 `form:"skip" binding:"min=0" default:"0"`
	Limit int64 `form:"limit" binding:"min=1,max=100" default:"10"`
}

// UserResponse represents the standard user response format
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
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
		Skip  int64 `json:"skip"`
		Limit int64 `json:"limit"`
	} `json:"meta"`
}

// UpdatePasswordRequest represents the request body for updating a password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// Add new handlers
func deleteCurrentUserHandler(ctx context.Context, _ any) (*MessageResponse, error) {
	// Get current user ID from context
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	err = db.DeleteUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error deleting user: %w", err)
	}
	return &MessageResponse{Message: "User deleted successfully"}, nil
}

func updatePasswordHandler(ctx context.Context, req UpdatePasswordRequest) (*MessageResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	// Get user with current password hash from DB
	user, err := db.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Verify current password against stored hash
	if !utils.CheckPasswordHash(req.CurrentPassword, user.HashedPassword) {
		return nil, middleware.NewBadRequestError("incorrect current password")
	}

	// Hash the new password
	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Update with new password hash
	user, err = db.UpdateUser(ctx, gen.UpdateUserParams{
		ID:             user.ID,
		HashedPassword: hashedNewPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating password: %w", err)
	}

	return &MessageResponse{
		Message: "Password updated successfully",
	}, nil
}

// Update handler functions
func registerUserHandler(ctx context.Context, req RegisterUserRequest) (*UserResponse, error) {
	db := ctx.Value("database").(*models.DB)
	// Check if user with this email already exists
	exists, err := db.IsUserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if exists {
		return nil, middleware.NewBadRequestError("user with this email already exists")
	}
	hashPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}
	user, err := db.CreateUser(ctx, gen.CreateUserParams{
		ID:             uuid.New(),
		Email:          req.Email,
		HashedPassword: hashPassword,
		FullName:       pgtype.Text{String: req.FullName, Valid: true},
		IsActive:       true,
		IsSuperuser:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName.String,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

// Update response maps in other handlers to include phone_number
func updateCurrentUserHandler(ctx context.Context, req UpdateUserMeRequest) (*UserResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}

	user, err := db.UpdateUser(ctx, gen.UpdateUserParams{
		ID:       id,
		FullName: pgtype.Text{String: req.FullName, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName.String,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

// Update getCurrentUserHandler response
func getCurrentUserHandler(ctx context.Context, _ any) (*UserResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID")
	}
	user, err := db.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName.String,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

// Update getUserHandler response
func getUserHandler(ctx context.Context, req GetUserRequest) (*UserResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	if !claims.IsSuperUser {
		return nil, middleware.NewForbiddenError("only superusers can get other users")
	}
	db := ctx.Value("database").(*models.DB)

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID format")
	}

	user, err := db.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName.String,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

// Update updateUserHandler to handle phone number
func updateUserHandler(ctx context.Context, req UpdateUserRequest) (*UserResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	if !claims.IsSuperUser {
		return nil, middleware.NewForbiddenError("only superusers can update other users")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid user ID format")
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Save updates
	user, err := db.UpdateUser(ctx, gen.UpdateUserParams{
		ID:             id,
		Email:          req.Email,
		FullName:       pgtype.Text{String: req.FullName, Valid: true},
		IsActive:       req.IsActive,
		IsSuperuser:    req.IsSuperuser,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName.String,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
	}, nil
}

// Update listUsersHandler response
func listUsersHandler(ctx context.Context, req ListUsersRequest) (*UsersResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	if !claims.IsSuperUser {
		return nil, middleware.NewForbiddenError("only superusers can list users")
	}
	db := ctx.Value("database").(*models.DB)

	// Set default values for pagination
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Skip < 0 {
		req.Skip = 0
	}

	users, err := db.ListUsers(ctx, gen.ListUsersParams{
		Limit:  req.Limit,
		Offset: req.Skip,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	// Convert to response format
	userList := make([]UserResponse, len(users))
	for i, user := range users {
		userList[i] = UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FullName:    user.FullName.String,
			IsActive:    user.IsActive,
			IsSuperuser: user.IsSuperuser,
			CreatedAt:   user.CreatedAt.Time,
			UpdatedAt:   user.UpdatedAt.Time,
		}
	}

	return &UsersResponse{
		Users: userList,
		Meta: struct {
			Skip  int64 `json:"skip"`
			Limit int64 `json:"limit"`
		}{
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	}, nil
}
