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

		logger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("trace_id", traceID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr), // cannot be true ip
			slog.String("user_agent", r.UserAgent()),
		)

		ctx := WithLogger(r.Context(), logger)
		ctx = WithRequestID(ctx, reqID)
		ctx = WithTraceID(ctx, traceID)

		// Response info
		rw := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		GetLoggerFromContext(ctx).Info("Request started")

		// Process request
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Log the end of request with metrics
		duration := time.Since(start)
		GetLoggerFromContext(ctx).Info("Request completed",
			slog.Int("status_code", rw.statusCode),
			slog.Int64("response_size", rw.bytesWritten),
			slog.Duration("duration", duration),
		)
	})
}
