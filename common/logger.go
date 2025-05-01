package common

import (
	"log/slog"
	"os"
)

var log *Logger

// Logger is a wrapper around slog.Logger
type Logger struct {
	slog.Logger
}

// Initialize sets up the logger
func Initialize(env string) {
	var handler slog.Handler

	if env == "production" {
		// JSON handler for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
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

	log = &Logger{
		Logger: *slog.New(handler),
	}
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if log == nil {
		Initialize(os.Getenv("ENV"))
	}
	return log
}
