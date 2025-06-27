package delivery

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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

func (ph *PostHandler) ListPublishedAndPaginatedPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_posts_handler")

	input, err := parseListPostQueryParams(r)
	if err != nil {
		log.Warn("Invalid query parameters", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	log.Info("Listing posts",
		slog.Int("page", input.Page),
		slog.Int("page_size", input.PageSize),
	)

	posts, err := ph.service.ListPublishedAndPaginatedPosts(ctx, input.Page, input.PageSize)
	if err != nil {
		log.Error("Failed to list posts via service",
			slog.Int("page", input.Page),
			slog.Int("page_size", input.PageSize),
			slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to retrieve posts")
		return
	}

	log.Info("Posts retrieved successfully", slog.Int("count", len(posts)))

	previews := make([]dto.PostPreviewResponse, len(posts))
	for i, post := range posts {
		previews[i] = dto.PostPreviewResponse{
			ID:          post.ID,
			Title:       post.Title,
			Description: post.Description,
			ImageID:     post.ImageID,
			PublishedAt: post.PublishedAt,
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"posts": previews,
		"pagination": map[string]int{
			"page":      input.Page,
			"page_size": input.PageSize,
		},
	})
}

func parseListPostQueryParams(r *http.Request) (dto.ListPostsInput, error) {
	input := dto.ListPostsInput{
		Page:     1,  // Default page
		PageSize: 10, // Default page size
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return input, fmt.Errorf("invalid page parameter: must be a number")
		}
		if p < 1 {
			return input, fmt.Errorf("invalid page parameter: must be greater than 0")
		}
		input.Page = p
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return input, fmt.Errorf("invalid page_size parameter: must be a number")
		}
		if ps < 1 || ps > 25 {
			return input, fmt.Errorf("invalid page_size parameter: must be between 1 and 25")
		}
		input.PageSize = ps
	}
	return input, nil
}
