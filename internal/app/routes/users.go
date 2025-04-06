package routes

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/repository"
)

// Request structs for users routes
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	FullName    string `json:"full_name" binding:"required"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

type UpdateUserRequest struct {
	ID          string `path:"id" binding:"required"`
	Email       string `json:"email" binding:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	IsActive    bool   `json:"is_active"`
	IsSuperuser bool   `json:"is_superuser"`
}

type GetUserRequest struct {
	ID string `path:"id" binding:"required"`
}

type ListUsersRequest struct {
	Skip  int `query:"skip"`
	Limit int `query:"limit"`
}

// RegisterUsersRoutes registers all user related routes
func RegisterUsersRoutes(h *server.Hertz) {
	usersGroup := h.Group("/api/v1/users")
	{
		usersGroup.POST("/",
			middleware.WithRequest(CreateUserRequest{}),
			middleware.WithResponse(createUserHandler))

		usersGroup.GET("/me",
			middleware.WithResponse(getCurrentUserHandler))

		usersGroup.GET("/:id",
			middleware.WithRequest(GetUserRequest{}),
			middleware.WithResponse(getUserHandler))

		usersGroup.PUT("/:id",
			middleware.WithRequest(UpdateUserRequest{}),
			middleware.WithResponse(updateUserHandler))

		usersGroup.GET("/",
			middleware.WithRequest(ListUsersRequest{}),
			middleware.WithResponse(listUsersHandler))
	}
}

// User handlers
func createUserHandler(ctx context.Context, req CreateUserRequest) (interface{}, error) {
	// Get the user repository from the context
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Check if user with this email already exists
	existingUser, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		// Return a specific error type that middleware can convert to 400
		return nil, middleware.NewBadRequestError("user with this email already exists")
	}

	// Create a new user model from the request
	user := &models.User{
		Email:       req.Email,
		Password:    req.Password, // Will be hashed in the repository
		FullName:    req.FullName,
		IsActive:    req.IsActive,
		IsSuperuser: req.IsSuperuser,
	}

	// Save the user to the database
	err = userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Return the created user (without password)
	return map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"full_name":    user.FullName,
		"is_active":    user.IsActive,
		"is_superuser": user.IsSuperuser,
		"created_at":   user.CreatedAt,
	}, nil
}

func getCurrentUserHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	// TODO: Get user ID from JWT token
	return nil, middleware.NewBadRequestError("not implemented: getCurrentUserHandler")
}

func getUserHandler(ctx context.Context, req GetUserRequest) (interface{}, error) {
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

	return map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"full_name":    user.FullName,
		"is_active":    user.IsActive,
		"is_superuser": user.IsSuperuser,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}, nil
}

func updateUserHandler(ctx context.Context, req UpdateUserRequest) (interface{}, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

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
		user.Password = req.Password
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

	return map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"full_name":    user.FullName,
		"is_active":    user.IsActive,
		"is_superuser": user.IsSuperuser,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}, nil
}

func listUsersHandler(ctx context.Context, req ListUsersRequest) (interface{}, error) {
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
	userList := make([]map[string]interface{}, len(users))
	for i, user := range users {
		userList[i] = map[string]interface{}{
			"id":           user.ID,
			"email":        user.Email,
			"full_name":    user.FullName,
			"is_active":    user.IsActive,
			"is_superuser": user.IsSuperuser,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		}
	}

	return map[string]interface{}{
		"users": userList,
		"meta": map[string]interface{}{
			"skip":  req.Skip,
			"limit": req.Limit,
		},
	}, nil
}
