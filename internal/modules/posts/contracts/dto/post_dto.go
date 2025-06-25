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
	Title     string `json:"title" validate:"required,min=2"`
	Content   string `json:"content" validate:"required"`
	AuthorID  string `json:"author_id" validate:"required,uuid4"`
	ImageID   string `json:"image_id" validate:"omitempty,uuid4"`
	Published bool   `json:"published"`
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
		Title:     cpr.Title,
		Content:   cpr.Content,
		AuthorID:  authorUUID,
		ImageID:   imageUUID,
		Published: cpr.Published,
	}, nil
}

// PostResponse represents the data returned when creating or fetching a post.
type PostResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Slug        string     `json:"slug"`
	AuthorID    string     `json:"author_id"`
	ImageID     *string    `json:"image_id,omitempty"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func FromPostModel(createdPost model.Post) PostResponse {
	return PostResponse{
		ID:       createdPost.ID.String(),
		Title:    createdPost.Title,
		Content:  createdPost.Content,
		Slug:     createdPost.Slug,
		AuthorID: createdPost.AuthorID.String(),
		ImageID: func() *string {
			if createdPost.ImageID != nil {
				str := createdPost.ImageID.String()
				return &str
			}
			return nil
		}(),
		Published:   createdPost.Published,
		PublishedAt: createdPost.PublishedAt,
		CreatedAt:   createdPost.CreatedAt,
		UpdatedAt:   createdPost.UpdatedAt,
	}
}
