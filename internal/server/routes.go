package server

import (
	"database/sql"
	"log"
	"net/http"

	categoryDelivery "github.com/Guizzs26/personal-blog/internal/modules/categories/delivery"
	categoryRepo "github.com/Guizzs26/personal-blog/internal/modules/categories/repository"
	categoryService "github.com/Guizzs26/personal-blog/internal/modules/categories/service"
	commentDelivery "github.com/Guizzs26/personal-blog/internal/modules/comments/delivery"
	commentRepo "github.com/Guizzs26/personal-blog/internal/modules/comments/repository"
	commentService "github.com/Guizzs26/personal-blog/internal/modules/comments/service"
	userDelivery "github.com/Guizzs26/personal-blog/internal/modules/identity/delivery"
	userRepository "github.com/Guizzs26/personal-blog/internal/modules/identity/repository"
	userService "github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	postDelivery "github.com/Guizzs26/personal-blog/internal/modules/posts/delivery"
	postRepo "github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
	postService "github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/Guizzs26/personal-blog/internal/server/handlers"
	"github.com/Guizzs26/personal-blog/pkg/cronx"
	"github.com/Guizzs26/personal-blog/pkg/jwtx"
)

func RegisterHTTPRoutes(mux *http.ServeMux, pgConn *sql.DB) {
	mux.HandleFunc("GET /health", handlers.HealthCheckHandler)

	// --- Users & Auth ---
	userRepo := userRepository.NewPostgresUserRepository(pgConn)
	userSvc := userService.NewUserService(userRepo)
	userHandler := userDelivery.NewUserHandler(*userSvc)
	refreshTokenRepo := userRepository.NewPostgresRefreshTokenRepository(pgConn)
	authService := userService.NewAuthService(userRepo, refreshTokenRepo)
	githubService := userService.SetupGitHubOAuth()
	authHandler := userDelivery.NewAuthHandler(*authService, *githubService)

	// Start cron jobs
	setupCron(authService)

	// --- Categories ---
	categoryRepo := categoryRepo.NewPostgresCategoryRepository(pgConn)
	categorySvc := categoryService.NewCategoryService(categoryRepo)
	categoryHandler := categoryDelivery.NewCategoryHandler(*categorySvc)

	// --- Posts ---
	postRepo := postRepo.NewPostgresPostRepository(pgConn)
	postSvc := postService.NewPostService(postRepo, categoryRepo)
	postHandler := postDelivery.NewPostHandler(*postSvc)

	// --- Comments ---
	commentRepo := commentRepo.NewPostgresCommentsRepository(pgConn)
	commentSvc := commentService.NewCommentService(commentRepo, postRepo)
	commentHandler := commentDelivery.NewCommentHandler(*commentSvc)

	// --- Category routes ---
	mux.Handle("POST /category", protectedRoute(categoryHandler.CreateCategoryHandler))
	mux.Handle("GET /category", protectedRoute(categoryHandler.ListCategoriesHandler))
	mux.Handle("PATCH /category/{id}", protectedRoute(categoryHandler.UpdateCategoryByIDHandler))
	mux.Handle("PATCH /category/{id}/toggle-active", protectedRoute(categoryHandler.ToggleCategoryActiveHandler))

	// --- Post routes ---
	mux.Handle("POST /post", protectedRoute(postHandler.CreatePostHandler))
	mux.Handle("GET /post", protectedRoute(postHandler.ListPostsHandler))
	mux.Handle("GET /post/{slug}", protectedRoute(postHandler.GetPostBySlugHandler))
	mux.Handle("PATCH /post/{id}/toggle-active", protectedRoute(postHandler.TogglePostActiveHandler))
	mux.Handle("PATCH /post/{id}", protectedRoute(postHandler.UpdatePostByIDHandler))
	mux.Handle("DELETE /post/{id}", protectedRoute(postHandler.DeletePostByIDHandler))

	// --- Comment routes ---
	mux.Handle("POST /comment", protectedRoute(commentHandler.CreateCommentHandler))
	mux.Handle("GET /post/{id}/comments", protectedRoute(commentHandler.ListPostCommentsHandler))

	// --- User routes ---
	mux.HandleFunc("POST /user", userHandler.CreateUserHandler)

	// --- Auth routes ---
	mux.HandleFunc("GET /auth/github/login", authHandler.GitHubLogin)
	mux.HandleFunc("GET /auth/github/callback", authHandler.GitHubCallback)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/refresh", authHandler.RefreshTokenHandler)
}

func protectedRoute(handler http.HandlerFunc) http.Handler {
	return jwtx.JWTAuthMiddleware(http.HandlerFunc(handler))
}

func setupCron(authService *userService.AuthService) {
	if err := cronx.StartCleanupCronJob(authService); err != nil {
		log.Fatalf("Failed to start cleanup cron job: %v", err)
	}
}
