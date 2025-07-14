package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

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
