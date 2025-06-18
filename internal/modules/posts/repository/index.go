package repository

import (
	"database/sql"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
)

type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new instance of PostRepository with the provided database connection.
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create inserts a new post into the database and returns the created post.
func (pr *PostRepository) Create(post model.Post) (*model.Post, error) {
	query := `
		INSERT INTO posts 
			(title, content, slug, author_id, image_id, published, published_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING 
			id, title, content, slug, author_id, image_id, 
			published, published_at, created_at, updated_at
	`

	var savedPost model.Post
	err := pr.db.QueryRow(
		query,
		post.Title,
		post.Content,
		post.Slug,
		post.AuthorID,
		post.ImageID,
		post.Published,
		post.PublishedAt,
	).Scan(
		&savedPost.ID,
		&savedPost.Title,
		&savedPost.Content,
		&savedPost.Slug,
		&savedPost.AuthorID,
		&savedPost.ImageID,
		&savedPost.Published,
		&savedPost.PublishedAt,
		&savedPost.CreatedAt,
		&savedPost.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert post: %w", err)
	}

	return &savedPost, nil
}
