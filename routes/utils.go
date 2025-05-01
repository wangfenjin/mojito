package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/wangfenjin/mojito/middleware"
)

// RegisterUtilRoutes registers all utility related routes
func RegisterUtilRoutes(r chi.Router) {
	r.Route("/api/v1/utils", func(r chi.Router) {
		r.Get("/health-check/", middleware.WithHandler(healthCheckHandler))
		r.Post("/test-email/", middleware.WithHandler(testEmailHandler))
	})
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
