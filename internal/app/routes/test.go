package routes

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/models/gen"
)

// RegisterTestRoutes registers test-related routes
func RegisterTestRoutes(r chi.Router) {
	r.Route("/api/v1/test", func(r chi.Router) {
		r.Delete("/cleanup", middleware.WithHandler(cleanupHandler))
		r.Get("/shutdown", middleware.WithHandler(shutdownHandler))
	})
}

// EmptyRequest represents an empty request
type EmptyRequest struct{}

func shutdownHandler(_ context.Context, _ EmptyRequest) (*MessageResponse, error) {
	defer os.Exit(0)
	return &MessageResponse{
		Message: "server shutting down",
	}, nil
}

func cleanupHandler(ctx context.Context, _ EmptyRequest) (*MessageResponse, error) {
	db := ctx.Value("database").(*models.DB)

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
