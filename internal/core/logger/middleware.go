package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := uuid.NewString()
		traceID := r.Header.Get("x-trace-id")
		if traceID == "" {
			traceID = reqID
		}

		// Basic logger with only IDs for internal use
		baseLogger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("trace_id", traceID),
		)

		// Complete logger for request/response logs only
		requestLogger := baseLogger.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		ctx := WithLogger(r.Context(), baseLogger)
		ctx = WithRequestID(ctx, reqID)
		ctx = WithTraceID(ctx, traceID)

		// Response info
		rw := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		requestLogger.Info("Request started")

		// Process request
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Log the end of request with metrics
		duration := time.Since(start)

		httpGroup := slog.Group("http",
			slog.Int("status", rw.statusCode),
			slog.Int64("size", rw.bytesWritten),
			slog.Duration("duration", duration))

		finalLogger := GetLoggerFromContext(ctx)

		switch {
		case rw.statusCode >= 500:
			finalLogger.Error("Request failed", httpGroup)
		case rw.statusCode >= 400:
			finalLogger.Warn("Client error", httpGroup)
		default:
			finalLogger.Info("Request completed", httpGroup)
		}
	})
}
