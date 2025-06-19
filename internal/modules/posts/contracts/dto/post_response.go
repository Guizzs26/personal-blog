package dto

import "time"

// PostResponse represents the data returned when creating or fetching a post.
type PostResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Slug        string    `json:"slug"`
	AuthorID    string    `json:"author_id"`
	ImageID     string    `json:"image_id"`
	Published   bool      `json:"published"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
