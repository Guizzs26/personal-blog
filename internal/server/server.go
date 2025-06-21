package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
)

func NewServer(pgConn *sql.DB) *http.Server {
	mux := http.NewServeMux()

	RegisterHTTPRoutes(mux, pgConn)

	handlerWithLogging := logger.LoggingMiddleware(mux)

	return &http.Server{
		Addr:              ":4444",
		Handler:           handlerWithLogging,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
}
