package delivery

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

	var req dto.CreatePostRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" || strings.TrimSpace(req.AuthorID) == "" || strings.TrimSpace(req.ImageID) == "" {
		http.Error(w, "title, content, author_id, and image_id are required", http.StatusBadRequest)
		return
	}

	authorUUID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		http.Error(w, "Invalid author_id", http.StatusBadRequest)
		return
	}

	imageUUID, err := uuid.Parse(req.ImageID)
	if err != nil {
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
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	res := dto.PostResponse{
		ID:        createdPost.ID.String(),
		Title:     createdPost.Title,
		Content:   createdPost.Content,
		Slug:      createdPost.Slug,
		AuthorID:  createdPost.AuthorID.String(),
		ImageID:   createdPost.ImageID.String(),
		Published: createdPost.Published,
		PublishedAt: func() time.Time {
			if createdPost.PublishedAt != nil {
				return *createdPost.PublishedAt
			}
			return time.Time{}
		}(),
		CreatedAt: createdPost.CreatedAt,
		UpdatedAt: createdPost.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
