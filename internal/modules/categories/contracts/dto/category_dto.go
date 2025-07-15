package dto

import (
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
)

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

// ToModel transforms a CreateCategoryRequest into a "domain" model.Category
func (ccr *CreateCategoryRequest) ToModel() model.Category {
	return model.Category{
		Name: ccr.Name,
	}
}

// CategoryFullResponse represents the complete data returned when fetching a post
// or when create/update a post
type CategoryFullResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToCategoryFullResponse(category *model.Category) CategoryFullResponse {
	return CategoryFullResponse{
		ID:        category.ID.String(),
		Name:      category.Name,
		Slug:      category.Slug,
		Active:    category.Active,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// PaginationParams represents basic pagination input parameters for paginated endpoints
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginationInfo contains metadata returned alongside paginated results
type PaginationInfo struct {
	Page        int  `json:"page"`
	PageSize    int  `json:"page_size"`
	TotalCount  int  `json:"total_count"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// PaginatedCategoriesResponse wraps a list of categories with pagination metadata
type PaginatedCategoriesResponse struct {
	Categories []CategoryFullResponse `json:"categories"`
	Pagination PaginationInfo         `json:"pagination"`
}

// NewPaginationInfo builds pagination metadata given the current page and total count
func NewPaginationInfo(page, pageSize, totalCount int) PaginationInfo {
	if totalCount < 0 {
		totalCount = 0
	}

	totalPages := 1
	if totalCount > 0 {
		totalPages = (totalCount + pageSize - 1) / pageSize
	}

	// Validate if the requested page exists
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	return PaginationInfo{
		Page:        page,
		PageSize:    pageSize,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		HasNext:     page < totalPages && totalCount > 0,
		HasPrevious: page > 1,
	}
}
