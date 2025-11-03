// Package logger provides structured logging using Go's slog package.
package logger

import (
	"log/slog"
	"os"

	"github.com/tyokyo320/rateflow/internal/infrastructure/config"
)

// New creates a new structured logger based on configuration.
func New(cfg config.LoggerConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Include source file and line number
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// WithContext adds common context fields to a logger.
func WithContext(logger *slog.Logger, service, version string) *slog.Logger {
	return logger.With(
		slog.String("service", service),
		slog.String("version", version),
	)
}

// WithRequest adds HTTP request fields to a logger.
func WithRequest(logger *slog.Logger, method, path, requestID string) *slog.Logger {
	return logger.With(
		slog.String("method", method),
		slog.String("path", path),
		slog.String("request_id", requestID),
	)
}
