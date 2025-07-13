package log

import (
	"log/slog"
	"os"
)

// New creates a new slog.Logger based on the provided configuration.
func New(lvl string) *slog.Logger {
	var level slog.Level
	switch lvl {
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	return logger
}
