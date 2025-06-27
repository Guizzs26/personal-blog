package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/google/uuid"
)

// CreatePostRequest represents the data required to create a new post.
type CreatePostRequest struct {
	Title       string `json:"title" validate:"required,min=2"`
	Content     string `json:"content" validate:"required"`
	Description string `json:"description" validate:"required,min=2,max=400"`
	AuthorID    string `json:"author_id" validate:"required,uuid4"`
	ImageID     string `json:"image_id" validate:"omitempty,uuid4"`
	Published   bool   `json:"published"`
}

func (cpr *CreatePostRequest) ToModel() (model.Post, error) {
	authorUUID, err := uuid.Parse(cpr.AuthorID)
	if err != nil {
		return model.Post{}, fmt.Errorf("failed to parse author_id to a valid uuid: %w", err)
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
		AuthorID:    authorUUID,
		ImageID:     imageUUID,
		Published:   cpr.Published,
	}, nil
}

// PostResponse represents the data returned when creating or fetching a post.
type PostResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Description string     `json:"description"`
	Slug        string     `json:"slug"`
	AuthorID    string     `json:"author_id"`
	ImageID     *string    `json:"image_id,omitempty"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func FromPostModel(post model.Post) PostResponse {
	return PostResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Content:     post.Content,
		Description: post.Description,
		Slug:        post.Slug,
		AuthorID:    post.AuthorID.String(),
		ImageID: func() *string {
			if post.ImageID != nil {
				str := post.ImageID.String()
				return &str
			}
			return nil
		}(),
		Published:   post.Published,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}
}

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type PaginationInfo struct {
	Page        int  `json:"page"`
	PageSize    int  `json:"page_size"`
	TotalCount  int  `json:"total_count"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

type PostPreviewResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageID     *uuid.UUID `json:"image_id,omitempty"`
	PublishedAt time.Time  `json:"published_at"`
}

type PaginatedPostsResponse struct {
	Posts      []PostPreviewResponse `json:"posts"`
	Pagination PaginationInfo        `json:"pagination"`
}

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
