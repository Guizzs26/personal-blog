package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerKey       contextKey = "logger"
	requestIDKey    contextKey = "request_id"
	traceIDKey      contextKey = "trace_id"
	ctxKeyUserAgent contextKey = "user_agent"
	ctxKeyIP        contextKey = "ip_address"
)

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ctxKeyIP, ip)
}

func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, ctxKeyUserAgent, ua)
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func GetRequestIDFromContext(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDKey).(string); ok {
		return rid
	}
	return ""
}

func GetTraceIDFromContext(ctx context.Context) string {
	if tid, ok := ctx.Value(traceIDKey).(string); ok {
		return tid
	}
	return ""
}

func GetIPAddressFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(ctxKeyIP).(string); ok {
		return ip
	}
	return ""
}

func GetUserAgentFromContext(ctx context.Context) string {
	if ua, ok := ctx.Value(ctxKeyUserAgent).(string); ok {
		return ua
	}
	return ""
}
