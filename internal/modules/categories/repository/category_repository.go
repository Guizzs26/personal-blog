package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

var ErrResourceNotFound = errors.New("resource not found")

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (cr *PostgresCategoryRepository) Create(ctx context.Context, category model.Category) (*model.Category, error) {
	query := `
		INSERT INTO categories
			(name, slug)
		VALUES
			($1, $2)
		RETURNING 
			id, name, slug, active, created_at, updated_at
	`

	var savedCategory model.Category
	err := cr.db.QueryRowContext(
		ctx, query, category.Name, category.Slug,
	).Scan(
		&savedCategory.ID,
		&savedCategory.Name,
		&savedCategory.Slug,
		&savedCategory.Active,
		&savedCategory.CreatedAt,
		&savedCategory.UpdatedAt,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: insert category: %v", err), 0)
	}

	return &savedCategory, nil
}

func (cr *PostgresCategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM categories WHERE slug = $1 AND active = true)
	`

	if err := cr.db.QueryRowContext(ctx, query, slug).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check slug existence by slug: %v", err), 0)
	}

	return exists, nil
}

func (cr *PostgresCategoryRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND active = true)
	`

	if err := cr.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check category exists by id: %v", err), 0)
	}

	return exists, nil
}

func (cr *PostgresCategoryRepository) ListActives(ctx context.Context, page, pageSize int) (*[]model.Category, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_active_repository")

	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, slug, active, created_at, updated_at
		FROM categories
		WHERE active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := cr.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: list active categories: %v", err), 0)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Active, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("repository: scan category row: %v", err), 0)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: iterate rows: %v", err), 0)
	}

	log.Debug("Listing active categories", slog.Int("page", page), slog.Int("page_size", pageSize))

	return &categories, nil
}

func (cr *PostgresCategoryRepository) CountActives(ctx context.Context) (int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("count_active_repository")

	var count int
	query := `
		SELECT COUNT(*)
		FROM categories
		WHERE active = true
	`

	if err := cr.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, xerrors.WithStackTrace(fmt.Errorf("repository: count active categories: %v", err), 0)
	}

	log.Debug("Counted active categories", slog.Int("count", count))
	return count, nil
}

func (cr *PostgresCategoryRepository) UpdateByID(ctx context.Context, id uuid.UUID, name, slug string) (*model.Category, error) {
	query := `
		UPDATE categories
		SET name = $1, slug = $2
		WHERE id = $3
		RETURNING id, name, slug, active, created_at, updated_at
	`

	var category model.Category
	err := cr.db.QueryRowContext(ctx, query, name, slug, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Active,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan updated category: %v", err), 0)
	}

	return &category, nil
}

func (cr *PostgresCategoryRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) (*model.Category, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("set_active_repository")

	query := `
		UPDATE categories
		SET active = $1,
				updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, slug, active, created_at, updated_at
	`

	row := cr.db.QueryRowContext(ctx, query, active, id)

	var category model.Category
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Active,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrResourceNotFound
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan category row: %v", err), 0)
	}

	status := "deactivated"
	if category.Active {
		status = "activated"
	}

	log.Info("Category status changed",
		slog.String("category_id", category.ID.String()),
		slog.String("status", status))
	return &category, nil
}
