package routes

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// RegisterUtilRoutes registers all utility related routes
func RegisterUtilRoutes(h *server.Hertz) {
	// Utils routes group
	utilsGroup := h.Group("/api/v1/utils")
	{
		// Health check endpoint
		utilsGroup.GET("/health-check/", healthCheckHandler)
		utilsGroup.POST("/test-email/", testEmailHandler)
	}
}

// Health check handler
func healthCheckHandler(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, true)
}

// Utils handlers
func testEmailHandler(ctx context.Context, c *app.RequestContext) {
	panic("Not implemented: testEmailHandler")
}
