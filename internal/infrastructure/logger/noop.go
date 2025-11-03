package logger

import (
	"context"
	"log/slog"
)

// NewNoop creates a no-op logger for testing.
func NewNoop() *slog.Logger {
	return slog.New(&noopHandler{})
}

type noopHandler struct{}

func (h *noopHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (h *noopHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *noopHandler) WithGroup(_ string) slog.Handler {
	return h
}
