package routes

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/wangfenjin/mojito/internal/app/middleware"
)

// RegisterUtilRoutes registers all utility related routes
func RegisterUtilRoutes(h *server.Hertz) {
	utilsGroup := h.Group("/api/v1/utils")
	{
		utilsGroup.GET("/health-check/",
			middleware.WithHandlerEmpty(healthCheckHandler))
		utilsGroup.POST("/test-email/",
			middleware.WithHandlerEmpty(testEmailHandler))
	}
}

// Update handler signatures
// Add response types
type HealthCheckResponse struct {
	Status bool `json:"status"`
}

func healthCheckHandler(ctx context.Context) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{Status: true}, nil
}

func testEmailHandler(ctx context.Context) (*MessageResponse, error) {
	panic("Not implemented: testEmailHandler")
}
