package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/google/uuid"
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
			(title, content, description, slug, category_id, author_id, image_id, published, published_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING 
			id, title, content, description, slug, category_id, author_id, image_id, 
			active, published, published_at, created_at, updated_at
	`

	var savedPost model.Post
	err := pr.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Description,
		post.Slug,
		post.CategoryID,
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
		&savedPost.CategoryID,
		&savedPost.AuthorID,
		&savedPost.ImageID,
		&savedPost.Active,
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
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE slug = $1 AND active = true)`

	if err := pr.db.QueryRowContext(ctx, query, slug).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check slug existence by slug: %v", err), 0)
	}

	log.Debug("Check Existence of a post with the given slug completed",
		slog.String("slug", slug),
		slog.Bool("exists", exists))

	return exists, nil
}

func (pr *PostgresPostRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("exists_by_id_repository")

	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)
	`

	if err := pr.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check slug existence by id: %v", err), 0)
	}

	log.Debug("Check Existence of a post with the given id completed",
		slog.String("id", id.String()),
		slog.Bool("exists", exists))

	return exists, nil
}

func (pr *PostgresPostRepository) IsInactiveByID(ctx context.Context, id uuid.UUID) (bool, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("is_inactive_by_id_repository")

	query := `SELECT active FROM posts WHERE id = $1`

	var active bool
	err := pr.db.QueryRowContext(ctx, query, id).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		log.Debug("Post not found for inactivation check", slog.String("id", id.String()))
		return false, ErrResourceNotFound
	}
	if err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: failed to check post status: %v", err), 0)
	}

	isInactive := !active

	return isInactive, nil
}

// ListPublished returns a paginated list of published posts,
// ordered by published_at descending. Only essential preview fields are fetched
func (pr *PostgresPostRepository) ListPublished(ctx context.Context, page, pageSize int, categorySlug *string) ([]model.PostPreview, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_published_repository")

	offset := (page - 1) * pageSize
	query := `
		SELECT 
			p.id, p.title, p.description, p.slug, p.image_id, p.published_at
		FROM posts p
		INNER JOIN categories c ON c.id = p.category_id
		WHERE 
				p.published = true AND p.active = true AND (c.slug = $3 OR $3 IS NULL)
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`

	categorySlugParam := sql.NullString{}
	if categorySlug != nil {
		categorySlugParam.Valid = true
		categorySlugParam.String = *categorySlug
	}

	rows, err := pr.db.QueryContext(ctx, query, pageSize, offset, categorySlugParam)
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

func (pr *PostgresPostRepository) CountPublished(ctx context.Context, categorySlug *string) (int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("count_published_repository")

	var count int
	query := `
		SELECT COUNT(*)
		FROM posts p
		INNER JOIN categories c ON c.id = p.category_id
		WHERE p.published = true AND p.active = true AND ($1::TEXT IS NULL OR c.slug = $1)
	`

	categorySlugParam := sql.NullString{}
	if categorySlug != nil {
		categorySlugParam.Valid = true
		categorySlugParam.String = *categorySlug
	}

	if err := pr.db.QueryRowContext(ctx, query, categorySlugParam).Scan(&count); err != nil {
		return 0, xerrors.WithStackTrace(fmt.Errorf("repository: count published posts: %v", err), 0)
	}

	log.Debug("Counted published posts", slog.Int("count", count))
	return count, nil
}

func (pr *PostgresPostRepository) FindPublishedBySlug(ctx context.Context, slug string) (*model.PostDetail, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("count_published_repository")

	query := `
		SELECT id, title, content, image_id, published_at
		FROM posts
		WHERE slug= $1 AND published = true AND active = true
		LIMIT 1
	`

	var post model.PostDetail
	err := pr.db.QueryRowContext(ctx, query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.ImageID,
		&post.PublishedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan post row: %v", err), 0)
	}

	log.Debug("Post found successfully", "slug", slug, "post_id", post.ID)
	return &post, nil
}

func (r *PostgresPostRepository) FindByIDIgnoreActive(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	const query = `
		SELECT id, title, content, description, slug, author_id, image_id, 
					 published, published_at, active, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var post model.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Description,
		&post.Slug,
		&post.AuthorID,
		&post.ImageID,
		&post.Published,
		&post.PublishedAt,
		&post.Active,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan post row: %v", err), 0)
	}

	return &post, nil
}

func (r *PostgresPostRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	const query = `
		SELECT id, title, content, description, slug, author_id, image_id, 
					 published, published_at, active, created_at, updated_at
		FROM posts
		WHERE id = $1 AND active = true
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var post model.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Description,
		&post.Slug,
		&post.AuthorID,
		&post.ImageID,
		&post.Published,
		&post.PublishedAt,
		&post.Active,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan post row: %v", err), 0)
	}

	return &post, nil
}

func (pr *PostgresPostRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("set_active_repository")

	query := `
		UPDATE posts
		SET active = $1,
		    updated_at = NOW()
		WHERE id = $2
		RETURNING id, title, content, description, slug, author_id, image_id, 
		published, published_at, active, created_at, updated_at
	`

	row := pr.db.QueryRowContext(ctx, query, active, id)

	var post model.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Description,
		&post.Slug,
		&post.AuthorID,
		&post.ImageID,
		&post.Published,
		&post.PublishedAt,
		&post.Active,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan post row: %v", err), 0)
	}

	status := "deactivated"
	if post.Active {
		status = "activated"
	}

	log.Info("Post status changed", slog.String("post_id", post.ID.String()), slog.String("status", status))
	return &post, nil
}

func (pr *PostgresPostRepository) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]any) (*model.Post, error) {
	setClauses := make([]string, 0, len(updates)+1)
	args := make([]any, 0, len(updates)+1)
	argPosition := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argPosition))
		args = append(args, value)
		argPosition++
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE posts
		SET %s
		WHERE id = $%d AND active = true
		RETURNING id, title, description, content, slug, active, published, published_at, 
					  image_id, author_id, created_at, updated_at
	`, strings.Join(setClauses, ", "), argPosition)

	args = append(args, id)
	row := pr.db.QueryRowContext(ctx, query, args...)

	var post model.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Description,
		&post.Content,
		&post.Slug,
		&post.Active,
		&post.Published,
		&post.PublishedAt,
		&post.ImageID,
		&post.AuthorID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan updated post: %v", err), 0)
	}

	return &post, nil
}

func (pr *PostgresPostRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	log := logger.GetLoggerFromContext(ctx).WithGroup("delete_post_by_id_repository")

	query := `DELETE FROM posts WHERE id = $1 AND active = false`

	r, err := pr.db.ExecContext(ctx, query, id)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("failed to execute delete query: %v", err), 0)
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("repository: could not check rows affected: %v", err), 0)
	}

	if rowsAffected == 0 {
		log.Debug("No post found to delete", slog.String("id", id.String()))
		return ErrResourceNotFound
	}

	log.Debug("Post deleted permanently", slog.String("id", id.String()))
	return nil
}
