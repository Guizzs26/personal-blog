package delivery

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/service"
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

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (ch *CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_category")

	req, err := httpx.Bind[dto.CreateCategoryRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	category := req.ToModel()
	createdCategory, err := ch.service.CreateCategory(ctx, category)
	if err != nil {
		log.Error("Failed to create category", slog.String("name", req.Name), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create post")
		return
	}

	log.Info("Category created", slog.String("id", createdCategory.ID.String()), slog.String("slug", createdCategory.Slug))

	res := dto.ToCategoryFullResponse(createdCategory)
	httpx.WriteJSON(w, http.StatusCreated, res)
}

func (ch *CategoryHandler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_categories")

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

	categories, totalCount, err := ch.service.ListActiveAndPaginatedCategories(ctx, input.Page, input.PageSize)
	if err != nil {
		log.Error("Failed to list categories", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to retrieve categories")
		return
	}

	catRes := make([]dto.CategoryFullResponse, len(*categories))
	for i, category := range *categories {
		catRes[i] = dto.ToCategoryFullResponse(&category)
	}

	res := dto.PaginatedCategoriesResponse{
		Categories: catRes,
		Pagination: dto.NewPaginationInfo(input.Page, input.PageSize, totalCount),
	}

	httpx.WriteJSON(w, http.StatusOK, res)
}

func (ch *CategoryHandler) UpdateCategoryByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("update_post_by_id")

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "category id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid category id format")
		return
	}

	req, err := httpx.Bind[dto.UpdateCategoryRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	category, err := ch.service.UpdateCategoryByID(ctx, id, req.Name)
	if errors.Is(err, service.ErrCategoryNotFound) {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "Category not found")
		return
	}
	if err != nil {
		log.Error("Failed to update category", slog.String("id", id.String()), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal error")
		return
	}

	log.Info("Category updated", slog.String("id", category.ID.String()))

	res := dto.ToCategoryFullResponse(category)
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (ch *CategoryHandler) ToggleCategoryActiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("toggle_category_active")

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "category id is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid category id format")
		return
	}

	inputData, err := httpx.Bind[struct {
		Active bool `json:"active"`
	}](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

	post, err := ch.service.SetCategoryActive(ctx, id, inputData.Active)
	if err != nil {
		log.Error("Failed to toggle category active status", slog.String("id", id.String()),
			slog.Bool("active", inputData.Active), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to update category status")
		return
	}

	log.Info("Category status updated", slog.String("id", id.String()), slog.Bool("active", inputData.Active))

	res := dto.ToCategoryFullResponse(post)
	httpx.WriteJSON(w, http.StatusOK, res)
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
