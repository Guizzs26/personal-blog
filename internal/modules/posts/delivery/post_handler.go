package delivery

import (
	"log/slog"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
)

// PostHandler handles HTTP requests related to posts
type PostHandler struct {
	service service.PostService
}

// NewPostHandler creates a new PostHandler with the given service
func NewPostHandler(service service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

// CreatePostHandler handles the creation of a new post via HTTP
func (ph *PostHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_post_handler")

	req, err := httpx.Bind[dto.CreatePostRequest](r)
	if err != nil {
		log.Warn("Invalid request payload", slog.Any("error", err))
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}

		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	log.Info("Creating post",
		slog.String("title", req.Title),
		slog.String("author_id", req.AuthorID),
		slog.Bool("published", req.Published),
	)
	post, err := req.ToModel()
	if err != nil {
		log.Warn("Failed to convert request to model", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid input data")
		return
	}

	createdPost, err := ph.service.CreatePost(ctx, post)
	if err != nil {
		log.Error("Failed to create post via service",
			slog.String("title", req.Title),
			slog.String("author_id", req.AuthorID),
			slog.Bool("published", req.Published),
			slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create post")
		return
	}

	log.Info("Post created successfully", slog.String("slug", createdPost.Slug))

	res := dto.FromPostModel(*createdPost)
	httpx.WriteJSON(w, http.StatusCreated, res)
}
