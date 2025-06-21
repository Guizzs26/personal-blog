package logger

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()
		traceID := reqID

		logger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("trace_id", traceID),
		)

		ctx := WithLogger(r.Context(), logger)
		ctx = WithRequestID(ctx, reqID)
		ctx = WithTraceID(ctx, traceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
