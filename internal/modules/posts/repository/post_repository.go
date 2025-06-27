package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/mdobak/go-xerrors"
)

var ErrResourceNotFound = errors.New("resource not found")

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
		INSERT INTO posts
			(title, content, description, slug, author_id, image_id, published, published_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING 
			id, title, content, description, slug, author_id, image_id, 
			published, published_at, created_at, updated_at
	`

	var savedPost model.Post
	err := pr.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Description,
		post.Slug,
		post.AuthorID,
		post.ImageID,
		post.Published,
		post.PublishedAt,
	).Scan(
		&savedPost.ID,
		&savedPost.Title,
		&savedPost.Content,
		&savedPost.Description,
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
	log := logger.GetLoggerFromContext(ctx).WithGroup("exists_by_slug_repository")

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

// ListPublished returns a paginated list of published posts,
// ordered by published_at descending. Only essential preview fields are fetched
func (pr *PostgresPostRepository) ListPublished(ctx context.Context, page, pageSize int) ([]model.PostPreview, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_published_repository")

	offset := (page - 1) * pageSize
	query := `
		SELECT id, title, description, slug, image_id, published_at
		FROM posts
		WHERE published = true
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := pr.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: list published posts: %v", err), 0)
	}
	defer rows.Close()

	var posts []model.PostPreview
	for rows.Next() {
		var p model.PostPreview
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Slug, &p.ImageID, &p.PublishedAt); err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("repository: scan post row: %v", err), 0)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: iterate rows: %v", err), 0)
	}

	log.Debug("Listing published posts", slog.Int("page", page), slog.Int("page_size", pageSize))

	return posts, nil
}

func (pr *PostgresPostRepository) CountPublished(ctx context.Context) (int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("count_published_repository")

	var count int
	query := `
		SELECT COUNT(*)
		FROM posts
		WHERE published = true
	`

	if err := pr.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, xerrors.WithStackTrace(fmt.Errorf("repository: count published posts: %v", err), 0)
	}

	log.Debug("Counted published posts", slog.Int("count", count))
	return count, nil
}

func (pr *PostgresPostRepository) FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("count_published_repository")

	query := `
		SELECT id, title, content, image_id, published_at
		FROM posts
		WHERE slug= $1 AND published = true
		LIMIT 1
	`

	log.Debug("executing query to find post by slug", "slug", slug)

	var post model.Post
	err := pr.db.QueryRowContext(ctx, query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.ImageID,
		&post.PublishedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("no post found with slug", "slug", slug)
		return nil, ErrResourceNotFound
	}
	if err != nil {
		log.Error("database error while finding post", "slug", slug, "error", err)
		return nil, xerrors.WithStackTrace(
			fmt.Errorf("failed to scan post row: %v", err), 0,
		)
	}

	log.Debug("post found successfully", "slug", slug, "post_id", post.ID)
	return &post, nil
}
