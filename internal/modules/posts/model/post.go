package model

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a model (database table) blog post
type Post struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Content     string     `json:"content" db:"content"`
	Description string     `json:"description" db:"description"`
	Slug        string     `json:"slug" db:"slug"`
	CategoryID  uuid.UUID  `json:"category_id" db:"category_id"`
	AuthorID    uuid.UUID  `json:"author_id" db:"author_id"`
	ImageID     *uuid.UUID `json:"image_id" db:"image_id"`
	Active      bool       `json:"active" db:"active"`
	Published   bool       `json:"published" db:"published"`
	PublishedAt *time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type PostPreview struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Slug        string     `json:"slug" db:"slug"`
	ImageID     *uuid.UUID `json:"image_id" db:"image_id"`
	PublishedAt time.Time  `json:"published_at" db:"published_at"`
}

type PostDetail struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Content     string     `json:"content" db:"content"`
	ImageID     *uuid.UUID `json:"image_id" db:"image_id"`
	PublishedAt time.Time  `json:"published_at" db:"published_at"`
}
