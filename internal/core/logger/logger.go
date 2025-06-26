package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// SetupLogger configures the global slog logger based on environment variables
func SetupLogger() {
	wrt := createWriterOutput()

	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if format == "" {
		format = "json"
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   level == slog.LevelDebug,
		ReplaceAttr: replaceAttr,
	}

	handler := newHandler(format, wrt, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// createLogOutput ensures log directory exists and returns a MultiWriter
func createWriterOutput() io.Writer {
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("failed to create ./logs directory: %v", err)
	}

	f, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open app.log file: %v", err)
	}

	return io.MultiWriter(os.Stdout, f)
}

// newHandler returns a slog.Handler based on format and output
func newHandler(frmt string, wrt io.Writer, opts *slog.HandlerOptions) slog.Handler {
	switch frmt {
	case "text":
		return slog.NewTextHandler(wrt, opts)
	case "json":
		return slog.NewJSONHandler(wrt, opts)
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format: %q. Using 'text' as default.\n", frmt)
		return slog.NewTextHandler(wrt, opts)
	}
}

// parseLogLevel maps string values to slog.Leveler with a fallback to Info
func parseLogLevel(l string) slog.Leveler {
	switch strings.ToLower(l) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Fprintf(os.Stderr, "Invalid log level: %q. Using 'info' as default.\n", l)
		return slog.LevelInfo
	}
}
