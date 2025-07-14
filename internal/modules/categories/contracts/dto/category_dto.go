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
