package delivery

import (
	"errors"
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

const (
	DefaultPage        = 1
	DefaultPageSize    = 10
	MaxPageSize        = 25
	MinPageAndPageSize = 1
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
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	post, err := req.ToModel()
	if err != nil {
		log.Warn("Failed to convert request to model", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid input data")
		return
	}

	createdPost, err := ph.service.CreatePost(ctx, post)
	if err != nil {
		log.Error("Failed to create post via service", slog.String("title", req.Title), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create post")
		return
	}

	log.Info("Post created successfully", slog.String("id", createdPost.ID.String()), slog.String("slug", createdPost.Slug))

	res := dto.ToPostFullResponse(*createdPost)
	httpx.WriteJSON(w, http.StatusCreated, res)
}

func (ph *PostHandler) ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_posts_handler")

	allowedParams := []string{"page", "page_size"}
	if err := validateAllowedQueryParams(r, allowedParams); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	input, err := parseListPostQueryParams(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	posts, totalCount, err := ph.service.ListPublishedAndPaginatedPosts(ctx, input.Page, input.PageSize)
	if err != nil {
		log.Error("Failed to list posts via service", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to retrieve posts")
		return
	}

	log.Info("Posts retrieved successfully", slog.Int("count", len(posts)), slog.Int("total", totalCount))

	previews := make([]dto.PostPreviewResponse, len(posts))
	for i, post := range posts {
		previews[i] = dto.ToPostPreviewResponse(post)
	}

	res := dto.PaginatedPostsResponse{
		Posts:      previews,
		Pagination: dto.NewPaginationInfo(input.Page, input.PageSize, totalCount),
	}

	httpx.WriteJSON(w, http.StatusOK, res)
}

func (ph *PostHandler) GetPostBySlugHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("get_post_by_slug_handler")

	slug := r.PathValue("slug")
	if slug == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "slug route parameter is required")
		return
	}

	post, err := ph.service.GetPublishedPostBySlug(ctx, slug)
	if errors.Is(err, service.ErrPostNotFound) {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "Post not found")
		return
	}

	if err != nil {
		log.Error("failed to get post by slug", slog.String("slug", slug), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal error")
		return
	}

	log.Info("Post retrieved successfully", slog.String("slug", slug))

	res := dto.ToPostDetailResponse(post)
	httpx.WriteJSON(w, 200, res)
}

func parseListPostQueryParams(r *http.Request) (dto.PaginationParams, error) {
	input := dto.PaginationParams{
		Page:     DefaultPage,     // Default page
		PageSize: DefaultPageSize, // Default page size
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			return input, fmt.Errorf("invalid page parameter: must be a number")
		}
		if p < MinPageAndPageSize {
			return input, fmt.Errorf("invalid page parameter: must be greater than 0")
		}
		input.Page = p
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return input, fmt.Errorf("invalid page_size parameter: must be a number")
		}
		if ps < MinPageAndPageSize || ps > MaxPageSize {
			return input, fmt.Errorf("invalid page_size parameter: must be between 1 and 25")
		}
		input.PageSize = ps
	}
	return input, nil
}

// validateAllowedQueryParams validates that all query parameters in the HTTP request
// are present in the allowed parameters whitelist. This function implements a defensive
// approach by rejecting any unknown parameters.
// Complexity Analysis:
//
//	Time: O(n + m) where n = len(allowed), m = number of query params in request
//	  - Set construction: O(n) - building the allowedSet map
//	  - Validation loop: O(m) - checking each query parameter
//	  - Map lookup: O(1) - constant time per parameter check
//	Space: O(n) - storage for allowedSet map with n entries using zero-byte struct{}
//
// Security: Implements whitelist validation to prevent parameter pollution attacks
func validateAllowedQueryParams(r *http.Request, allowed []string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, k := range allowed {
		allowedSet[k] = struct{}{}
	}

	for key := range r.URL.Query() {
		if _, ok := allowedSet[key]; !ok {
			return fmt.Errorf("query parameter '%s' is not allowed", key)
		}
	}

	return nil
}
