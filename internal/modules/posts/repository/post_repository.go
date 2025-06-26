package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/mdobak/go-xerrors"
)

// PostgresPostRepository handles database operations related to posts
type PostgresPostRepository struct {
	db *sql.DB
}

// NewPostgresPostRepository creates a new instance of PostRepository with the provided database connection
func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// Create inserts a new post into the database and returns the saved record
func (pr *PostgresPostRepository) Create(ctx context.Context, post model.Post) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_post_repository	")

	query := `
		INSERT INTO postss
			(title, content, slug, author_id, image_id, published, published_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7)
		RETURNING 
			id, title, content, slug, author_id, image_id, 
			published, published_at, created_at, updated_at
	`

	var savedPost model.Post
	err := pr.db.QueryRowContext(
		ctx,
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
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: insert post: %v", err), 0)
	}

	log.Debug("Post inserted successfully in repository",
		slog.String("post_id", savedPost.ID.String()),
		slog.String("slug", savedPost.Slug))

	return &savedPost, nil
}

// ExistsBySlug checks if a post with the given slug already exists
func (pr *PostgresPostRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	log := logger.GetLoggerFromContext(ctx)

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1)`

	if err := pr.db.QueryRowContext(ctx, query, slug).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check slug existence: %v", err), 0)
	}

	log.Debug("Slug existence check completed",
		slog.String("slug", slug),
		slog.Bool("exists", exists))

	return exists, nil
}
