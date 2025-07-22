package logger

import (
	"log/slog"
	"net"
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

func getRealIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := uuid.NewString()
		traceID := r.Header.Get("x-trace-id")
		if traceID == "" {
			traceID = reqID
		}

		ip := getRealIP(r)
		ua := r.UserAgent()

		// Logger base para uso interno (com RequestID e TraceID)
		baseLogger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("trace_id", traceID),
		)

		// Logger completo com dados de request
		requestLogger := baseLogger.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", ip),
			slog.String("user_agent", ua),
		)

		// Injeta contexto com logger e identificadores
		ctx := r.Context()
		ctx = WithLogger(ctx, baseLogger)
		ctx = WithRequestID(ctx, reqID)
		ctx = WithTraceID(ctx, traceID)
		ctx = WithIPAddress(ctx, ip)
		ctx = WithUserAgent(ctx, ua)

		rw := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		requestLogger.Info("Request started")

		// Processa a requisição
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Finaliza com métricas
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
