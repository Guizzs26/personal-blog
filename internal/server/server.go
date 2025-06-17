package server

import (
	"database/sql"
	"net/http"
	"time"
)

func NewServer(pgConn *sql.DB) *http.Server {
	mux := http.NewServeMux()

	RegisterHTTPRoutes(mux, pgConn)

	return &http.Server{
		Addr:              ":4444",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
}
