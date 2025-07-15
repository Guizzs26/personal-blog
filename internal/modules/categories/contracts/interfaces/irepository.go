package interfaces

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
	"github.com/google/uuid"
)

type ICategoryRepository interface {
	Create(ctx context.Context, category model.Category) (*model.Category, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ListActives(ctx context.Context, page, pageSize int) (*[]model.Category, error)
	CountActives(ctx context.Context) (int, error)
	UpdateByID(ctx context.Context, id uuid.UUID, name, slug string) (*model.Category, error)
}
