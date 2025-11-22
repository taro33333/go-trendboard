package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/yourname/go-trendboard/internal/config"
)

// NewLogger creates and returns a new slog.Logger based on the provided configuration.
func NewLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource adds the source code position to the log entry.
		// This is useful for debugging but can have a small performance impact.
		AddSource: level == slog.LevelDebug,
	}

	// In a real production environment, you might choose between JSONHandler and TextHandler
	// based on an environment variable or another configuration.
	// JSONHandler is generally better for machine processing (e.g., log aggregation systems).
	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler)
}
