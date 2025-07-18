package server

import (
	"database/sql"
	"net/http"

	categoryDelivery "github.com/Guizzs26/personal-blog/internal/modules/categories/delivery"
	categoryRepo "github.com/Guizzs26/personal-blog/internal/modules/categories/repository"
	categoryService "github.com/Guizzs26/personal-blog/internal/modules/categories/service"
	userDelivery "github.com/Guizzs26/personal-blog/internal/modules/identity/delivery"
	userRepo "github.com/Guizzs26/personal-blog/internal/modules/identity/repository"
	userService "github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/delivery"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/Guizzs26/personal-blog/internal/server/handlers"
)

func RegisterHTTPRoutes(mux *http.ServeMux, pgConn *sql.DB) {
	mux.HandleFunc("GET /health", handlers.HealthCheckHandler)

	// Users module
	userRepo := userRepo.NewPostgresUserRepository(pgConn)
	userService := userService.NewUserService(userRepo)
	userHandler := userDelivery.NewUserHandler(*userService)

	// Category module
	categoryRepo := categoryRepo.NewPostgresCategoryRepository(pgConn)
	categoryService := categoryService.NewCategoryService(categoryRepo)
	categoryHandler := categoryDelivery.NewCategoryHandler(*categoryService)

	// Posts module
	postRepo := repository.NewPostgresPostRepository(pgConn)
	postService := service.NewPostService(postRepo, categoryRepo)
	postHandler := delivery.NewPostHandler(*postService)

	mux.HandleFunc("POST /category", categoryHandler.CreateCategoryHandler)
	mux.HandleFunc("GET /category", categoryHandler.ListCategoriesHandler)
	mux.HandleFunc("PATCH /category/{id}", categoryHandler.UpdateCategoryByIDHandler)
	mux.HandleFunc("PATCH /category/{id}/toggle-active", categoryHandler.ToggleCategoryActiveHandler)

	mux.HandleFunc("POST /post", postHandler.CreatePostHandler)
	mux.HandleFunc("GET /post", postHandler.ListPostsHandler)
	mux.HandleFunc("GET /post/{slug}", postHandler.GetPostBySlugHandler)
	mux.HandleFunc("PATCH /post/{id}/toggle-active", postHandler.TogglePostActiveHandler)
	mux.HandleFunc("PATCH /post/{id}", postHandler.UpdatePostByIDHandler)
	mux.HandleFunc("DELETE /post/{id}", postHandler.DeletePostByIDHandler)

	mux.HandleFunc("POST /user", userHandler.CreateUserHandler)
}
