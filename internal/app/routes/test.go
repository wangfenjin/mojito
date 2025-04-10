package routes

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/repository"
)

// RegisterTestRoutes registers test-related routes
func RegisterTestRoutes(h *gin.Engine) {
	testGroup := h.Group("/api/v1/test")
	{
		testGroup.DELETE("/cleanup",
			middleware.WithHandler(cleanupHandler))
	}
}

type EmptyRequest struct{}

func cleanupHandler(ctx context.Context, _ EmptyRequest) (*MessageResponse, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	// cleanup item first, because item has foreign key constraint with user
	if err := itemRepo.CleanupTestData(ctx); err != nil {
		return nil, fmt.Errorf("error cleaning up item data: %w", err)
	}

	if err := userRepo.CleanupTestData(ctx); err != nil {
		return nil, fmt.Errorf("error cleaning up user data: %w", err)
	}

	return &MessageResponse{
		Message: "Test data cleaned up",
	}, nil
}
