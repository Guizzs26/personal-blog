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
}
