package delivery

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_post")

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
		log.Error("Failed to create post", slog.String("title", req.Title), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create post")
		return
	}

	log.Info("Post created", slog.String("id", createdPost.ID.String()), slog.String("slug", createdPost.Slug))

	res := dto.ToPostFullResponse(createdPost)
	httpx.WriteJSON(w, http.StatusCreated, res)
}

func (ph *PostHandler) ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_posts")

	allowedParams := []string{"page", "page_size", "category_slug"}
	if err := validateAllowedQueryParams(r, allowedParams); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	input, err := parseListPostQueryParams(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	posts, totalCount, err := ph.service.ListPublishedAndPaginatedPosts(ctx, input.Page, input.PageSize, input.CategorySlug)
	if err != nil {
		log.Error("Failed to list posts", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to retrieve posts")
		return
	}

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
	log := logger.GetLoggerFromContext(ctx).WithGroup("get_post_by_slug")

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
		log.Error("Failed to get post by slug", slog.String("slug", slug), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal error")
		return
	}

	res := dto.ToPostDetailResponse(post)
	httpx.WriteJSON(w, 200, res)
}

func (ph *PostHandler) TogglePostActiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("toggle_post_active")

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "post id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid post id format")
		return
	}

	inputData, err := httpx.Bind[struct {
		Active bool `json:"active"`
	}](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

	post, err := ph.service.SetPostActive(ctx, id, inputData.Active)
	if err != nil {
		log.Error("Failed to toggle post active status", slog.String("id", id.String()), slog.Bool("active", inputData.Active), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to update post status")
		return
	}

	log.Info("Post status updated", slog.String("id", id.String()), slog.Bool("active", inputData.Active))

	res := dto.ToPostFullResponse(post)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (ph *PostHandler) UpdatePostByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("update_post_by_id")

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "post id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid post id format")
		return
	}

	req, err := httpx.Bind[dto.UpdatePostRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	updates, err := req.ToUpdateMap()
	if err != nil {
		log.Warn("Failed to convert request to update map", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}

	post, err := ph.service.UpdatePostByID(ctx, id, updates)
	if errors.Is(err, service.ErrPostNotFound) {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "Post not found")
		return
	}
	if err != nil {
		log.Error("Failed to update post", slog.String("id", id.String()), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal error")
		return
	}

	log.Info("Post updated", slog.String("id", post.ID.String()))

	res := dto.ToPostFullResponse(post)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (ph *PostHandler) DeletePostByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("delete_post_by_id_handler")

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, 400, httpx.ErrorCodeBadRequest, "post id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid post id format")
		return
	}

	err = ph.service.DeletePostByID(ctx, id)
	if errors.Is(err, service.ErrPostIsActive) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Active post cannot be deleted")
		return
	}
	if errors.Is(err, service.ErrPostNotFound) {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "Post not found")
		return
	}
	if err != nil {
		log.Error("Failed to delete post", slog.String("id", id.String()), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal error")
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, nil)
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

	var categorySlug *string
	if slug := strings.TrimSpace(r.URL.Query().Get("category_slug")); slug != "" {
		categorySlug = &slug
	}

	input.CategorySlug = categorySlug
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
