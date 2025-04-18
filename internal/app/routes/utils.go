package routes

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/wangfenjin/mojito/internal/app/middleware"
)

// RegisterUtilRoutes registers all utility related routes
func RegisterUtilRoutes(r *gin.Engine) {
	utilsGroup := r.Group("/api/v1/utils")
	{
		utilsGroup.GET("/health-check/",
			middleware.WithHandler(healthCheckHandler))
		utilsGroup.POST("/test-email/",
			middleware.WithHandler(testEmailHandler))
	}
}

// HealthCheckResponse is the response for the health check
type HealthCheckResponse struct {
	Status bool `json:"status"`
}

func healthCheckHandler(_ context.Context, _ any) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{Status: true}, nil
}

func testEmailHandler(_ context.Context, _ any) (*MessageResponse, error) {
	panic("Not implemented: testEmailHandler")
}
