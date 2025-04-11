package logger

import (
	"context"
	"log/slog"
	"os"
)

var log *logger

type logger struct {
	slog.Logger
}

// Initialize sets up the logger
func Initialize(env string) {
	var handler slog.Handler

	if env == "production" {
		// JSON handler for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   "timestamp",
						Value: a.Value,
					}
				}
				if a.Key == slog.SourceKey {
					return slog.Attr{
						Key:   "src",
						Value: a.Value,
					}
				}
				return a
			},
		})
	} else {
		// Text handler with colors for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	log = &logger{
		Logger: *slog.New(handler),
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logger {
	if log == nil {
		Initialize(os.Getenv("ENV"))
	}
	return log
}

// WithContext adds request context information to the logger
func WithContext(ctx context.Context) *slog.Logger {
	return GetLogger().With(
		slog.String("request_id", GetRequestID(ctx)),
	)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}
