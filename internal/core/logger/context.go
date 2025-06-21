package logger

import (
	"context"
	"log/slog"
)

// contextKey is a custom type to prevent key collisions in context.
type contextKey string

// Keys used to store/retrieve values from context.
const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
)

// WithLogger stores a structured logger in the context
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// GetLoggerFromContext retrieves the logger from the context
// If not found, it return slog.Default()
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}

// WithRequestID stores the request ID in the context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestIDFromContext retrieves the request ID from the context
// Returns an empty string if not found
func GetRequestIDFromContext(ctx context.Context) string {
	rid, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return rid
}

// WithTraceID stores the trace ID in the context.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// GetTraceIDFromContext retrieves the trace ID from the context
// Returns an empty string if not found
func GetTraceIDFromContext(ctx context.Context) string {
	tid, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return ""
	}
	return tid
}
