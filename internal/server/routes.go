package server

import (
	"database/sql"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/server/handlers"
)

func RegisterHTTPRoutes(mux *http.ServeMux, pgConn *sql.DB) {
	mux.HandleFunc("GET /health", handlers.HealthCheckHandler)
}
