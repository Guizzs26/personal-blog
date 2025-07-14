package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/google/uuid"
)

// CreatePostRequest represents the payload used to create a new blog post
type CreatePostRequest struct {
	Title       string `json:"title" validate:"required,min=2"`
	Content     string `json:"content" validate:"required"`
	Description string `json:"description" validate:"required,min=2,max=400"`
	CategoryID  string `json:"category_id" validate:"required,uuid4"`
	AuthorID    string `json:"author_id" validate:"required,uuid4"`
	ImageID     string `json:"image_id" validate:"omitempty,uuid4"`
	Published   bool   `json:"published"`
}

// ToModel transforms a CreatePostRequest into a "domain" model.Post
func (cpr *CreatePostRequest) ToModel() (model.Post, error) {
	authorUUID, err := uuid.Parse(cpr.AuthorID)
	if err != nil {
		return model.Post{}, fmt.Errorf("failed to parse author_id to a valid uuid: %w", err)
	}

	categoryUUID, err := uuid.Parse(cpr.CategoryID)
	if err != nil {
		return model.Post{}, fmt.Errorf("failed to parse category_id to a valid uuid: %w", err)
	}

	var imageUUID *uuid.UUID
	if strings.TrimSpace(cpr.ImageID) != "" {
		parsedImageID, err := uuid.Parse(cpr.ImageID)
		if err != nil {
			return model.Post{}, fmt.Errorf("failed to parse image_id to a valid uuid: %w", err)
		}
		imageUUID = &parsedImageID
	}

	return model.Post{
		Title:       cpr.Title,
		Content:     cpr.Content,
		Description: cpr.Description,
		CategoryID:  categoryUUID,
		AuthorID:    authorUUID,
		ImageID:     imageUUID,
		Published:   cpr.Published,
	}, nil
}

// PostFullResponse represents the complete data returned when fetching a post
// or when create/update a post
type PostFullResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Description string     `json:"description"`
	Slug        string     `json:"slug"`
	CategoryID  string     `json:"category_id"`
	AuthorID    string     `json:"author_id"`
	ImageID     *string    `json:"image_id"`
	Active      bool       `json:"active"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToPostFullResponse converts a "domain" model.Post into a PostFullResponse DTO
func ToPostFullResponse(post *model.Post) PostFullResponse {
	var imageID *string
	if post.ImageID != nil {
		id := post.ImageID.String()
		imageID = &id
	}

	return PostFullResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Content:     post.Content,
		Description: post.Description,
		Slug:        post.Slug,
		CategoryID:  post.CategoryID.String(),
		AuthorID:    post.AuthorID.String(),
		ImageID:     imageID,
		Active:      post.Active,
		Published:   post.Published,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}
}

// UpdatePostRequest represents the payload to update a blog post.
// All fields are optional (pointers) to allow partial updates
type UpdatePostRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=2"`
	Content     *string `json:"content" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty,min=2,max=400"`
	AuthorID    *string `json:"author_id" validate:"omitempty,uuid4"`
	ImageID     *string `json:"image_id" validate:"omitempty,uuid4"`
	Published   *bool   `json:"published"`
}

func (upr *UpdatePostRequest) ToUpdateMap() (map[string]any, error) {
	updates := map[string]any{}

	if upr.Title != nil {
		updates["title"] = *upr.Title
	}

	if upr.Content != nil {
		updates["content"] = *upr.Content
	}

	if upr.Description != nil {
		updates["description"] = *upr.Description
	}

	if upr.AuthorID != nil {
		authorUUID, err := uuid.Parse(*upr.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("invalid author_id format: %v", err)
		}
		updates["author_id"] = authorUUID
	}

	if upr.ImageID != nil {
		ImageUUID, err := uuid.Parse(*upr.ImageID)
		if err != nil {
			return nil, fmt.Errorf("invalid author_id format: %v", err)
		}
		updates["image_id"] = ImageUUID
	}

	if upr.Published != nil {
		updates["published"] = *upr.Published
	}
	return updates, nil
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

// PostPreviewResponse is a lightweight post representation used in list views
type PostPreviewResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	ImageID     *string   `json:"image_id,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// ToPostFullResponse converts a model.PostPreview into a PostPreviewResponse DTO
func ToPostPreviewResponse(post model.PostPreview) PostPreviewResponse {
	var imageID *string
	if post.ImageID != nil {
		id := post.ImageID.String()
		imageID = &id
	}

	return PostPreviewResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Description: post.Description,
		Slug:        post.Slug,
		ImageID:     imageID,
		PublishedAt: post.PublishedAt,
	}
}

// PaginatedPostsResponse wraps a list of post previews with pagination metadata
type PaginatedPostsResponse struct {
	Posts      []PostPreviewResponse `json:"posts"`
	Pagination PaginationInfo        `json:"pagination"`
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

// PostDetailResponse represents a detailed but minimal view of a single post.
// Typically used for single post retrieval (GET /posts/{slug}).
type PostDetailResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	ImageID     *string   `json:"image_id"`
	PublishedAt time.Time `json:"published_at"`
}

// ToPostDetailResponse converts a model.PostDetail into a PostDetailResponse DTO
func ToPostDetailResponse(post *model.PostDetail) PostDetailResponse {
	var imageID *string
	if post.ImageID != nil {
		id := post.ImageID.String()
		imageID = &id
	}

	return PostDetailResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Content:     post.Content,
		ImageID:     imageID,
		PublishedAt: post.PublishedAt,
	}
}
