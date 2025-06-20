package logger

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogger configures the global slog logger based on environment variables
func SetupLogger() {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	format := os.Getenv("LOG_FORMAT")

	var handler slog.Handler

	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// parseLogLevel maps string values to slog.Leveler with a fallback to Info.
func parseLogLevel(l string) slog.Leveler {
	switch strings.ToLower(l) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
