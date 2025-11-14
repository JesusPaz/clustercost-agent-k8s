package logging

import (
	"log/slog"
	"os"
)

// New creates a JSON slog logger with the supplied level string.
func New(level string) *slog.Logger {
	lvl := parseLevel(level)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}

func parseLevel(level string) slog.Leveler {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
