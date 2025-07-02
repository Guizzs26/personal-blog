package server

import (
	"database/sql"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/delivery"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/Guizzs26/personal-blog/internal/server/handlers"
)

func RegisterHTTPRoutes(mux *http.ServeMux, pgConn *sql.DB) {
	mux.HandleFunc("GET /health", handlers.HealthCheckHandler)

	// Posts module
	postRepo := repository.NewPostgresPostRepository(pgConn)
	postService := service.NewPostService(postRepo)
	postHandler := delivery.NewPostHandler(*postService)

	mux.HandleFunc("POST /post", postHandler.CreatePostHandler)
	mux.HandleFunc("GET /post", postHandler.ListPostsHandler)
	mux.HandleFunc("GET /post/{slug}", postHandler.GetPostBySlugHandler)
	mux.HandleFunc("PATCH /post/{id}/toggle-active", postHandler.TogglePostActiveHandler)
	mux.HandleFunc("PATCH /post/{id}", postHandler.UpdatePostByIDHandler)
	mux.HandleFunc("DELETE /post/{id}", postHandler.DeletePostByIDHandler)
}
