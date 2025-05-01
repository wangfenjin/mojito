package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/models/gen"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

// RegisterTestRoutes registers test-related routes
func RegisterTestRoutes(r chi.Router) {
	r.Route("/api/v1/test", func(r chi.Router) {
		r.Delete("/cleanup", middleware.WithHandler(cleanupHandler))
		r.Get("/shutdown", middleware.WithHandler(shutdownHandler))
		r.Post("/superuser", middleware.WithHandler(createSuperUserHandler))
	})
}

// EmptyRequest represents an empty request
type EmptyRequest struct{}

// CreateSuperUserRequest represents the request body for creating a super user
type CreateSuperUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

func createSuperUserHandler(ctx context.Context, req CreateSuperUserRequest) (*MessageResponse, error) {
	db := models.GetDB()

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	if err := db.WithTx(ctx, func(q *gen.Queries) error {
		// Check if email already exists
		exists, err := q.IsUserEmailExists(ctx, req.Email)
		if err != nil {
			return fmt.Errorf("error checking email existence: %w", err)
		}
		if exists {
			return middleware.NewBadRequestError("email already exists")
		}

		// Create super user
		_, err = q.CreateUser(ctx, gen.CreateUserParams{
			ID:             uuid.New(),
			Email:          req.Email,
			HashedPassword: hashedPassword,
			IsActive:       true,
			IsSuperuser:    true,
			FullName:       pgtype.Text{String: req.FullName, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("error creating super user: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &MessageResponse{
		Message: "Super user created successfully",
	}, nil
}

func shutdownHandler(_ context.Context, _ EmptyRequest) (*MessageResponse, error) {
	defer os.Exit(0)
	return &MessageResponse{
		Message: "server shutting down",
	}, nil
}

func cleanupHandler(ctx context.Context, _ EmptyRequest) (*MessageResponse, error) {
	db := models.GetDB()

	if err := db.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.CleanupItems(ctx); err != nil {
			return fmt.Errorf("error deleting items: %w", err)
		}
		return q.CleanupUsers(ctx)
	}); err != nil {
		return nil, fmt.Errorf("error cleaning up test data: %w", err)
	}

	return &MessageResponse{
		Message: "Test data cleaned up",
	}, nil
}
