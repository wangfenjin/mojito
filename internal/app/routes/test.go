package routes

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/repository"
)

// RegisterTestRoutes registers test-related routes
func RegisterTestRoutes(h *server.Hertz) {
	testGroup := h.Group("/api/v1/test")
	{
		testGroup.DELETE("/cleanup", middleware.WithResponse(cleanupHandler))
	}
}

type EmptyRequest struct{}

func cleanupHandler(ctx context.Context, _ EmptyRequest) (interface{}, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Clean up test data
	err := userRepo.CleanupTestData(ctx)
	if err != nil {
		return nil, fmt.Errorf("error cleaning up test data: %w", err)
	}

	return map[string]string{"message": "Test data cleaned up"}, nil
}
