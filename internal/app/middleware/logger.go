package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
)

// LoggerMiddleware logs HTTP request/response details
func LoggerMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())

		// Generate request ID
		requestID := uuid.New().String()
		ctx = context.WithValue(ctx, "request_id", requestID)

		// Add request ID to response headers
		c.Response.Header.Set("X-Request-ID", requestID)

		c.Next(ctx)

		// Log request details
		logger.GetLogger().Info("HTTP Request",
			slog.String("method", string(c.Request.Method())),
			slog.String("path", path),
			slog.String("query", string(c.Request.URI().QueryString())),
			slog.Int("status", c.Response.StatusCode()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", string(c.UserAgent())),
			slog.String("request_id", requestID),
		)
	}
}
