package dto

// CreatePostRequest represents the data required to create a new post.
type CreatePostRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Slug      string `json:"slug"`
	AuthorID  string `json:"author_id"`
	ImageID   string `json:"image_id"`
	Published bool   `json:"published"`
}
