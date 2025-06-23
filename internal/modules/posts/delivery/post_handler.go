package delivery

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/google/uuid"
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

	var req dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Invalid request body", slog.Any("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Debug("Request payload received",
		slog.String("title", req.Title),
		slog.String("authod_id", req.AuthorID),
		slog.Bool("published", req.Published),
	)

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" || strings.TrimSpace(req.AuthorID) == "" || strings.TrimSpace(req.ImageID) == "" {
		log.Warn("Required fields missing",
			slog.Bool("title_empty", strings.TrimSpace(req.Title) == ""),
			slog.Bool("content_empty", strings.TrimSpace(req.Content) == ""),
			slog.Bool("author_id_empty", strings.TrimSpace(req.AuthorID) == ""),
			slog.Bool("image_id_empty", strings.TrimSpace(req.ImageID) == ""))

		http.Error(w, "title, content, author_id, and image_id are required", http.StatusBadRequest)
		return
	}

	authorUUID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		log.Warn("Invalid author_id", slog.String("author_id", req.AuthorID))
		http.Error(w, "Invalid author_id", http.StatusBadRequest)
		return
	}

	imageUUID, err := uuid.Parse(req.ImageID)
	if err != nil {
		log.Warn("Invalid image_id", slog.String("image_id", req.ImageID))
		http.Error(w, "Invalid image_id", http.StatusBadRequest)
		return
	}

	post := model.Post{
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  authorUUID,
		ImageID:   imageUUID,
		Published: req.Published,
	}

	createdPost, err := ph.service.CreatePost(ctx, post)
	if err != nil {
		log.Error("Service error", slog.Any("error", err))
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	res := dto.PostResponse{
		ID:          createdPost.ID.String(),
		Title:       createdPost.Title,
		Content:     createdPost.Content,
		Slug:        createdPost.Slug,
		AuthorID:    createdPost.AuthorID.String(),
		ImageID:     createdPost.ImageID.String(),
		Published:   createdPost.Published,
		PublishedAt: createdPost.PublishedAt,
		CreatedAt:   createdPost.CreatedAt,
		UpdatedAt:   createdPost.UpdatedAt,
	}

	log.Info("Post created", slog.String("slug", createdPost.Slug))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
