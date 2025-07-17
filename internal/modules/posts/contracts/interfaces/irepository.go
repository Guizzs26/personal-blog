package interfaces

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/google/uuid"
)

type IPostRepository interface {
	Create(ctx context.Context, post model.Post) (*model.Post, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ListPublished(ctx context.Context, page, pageSize int, categorySlug *string) ([]model.PostPreview, error)
	CountPublished(ctx context.Context, categorySlug *string) (int, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*model.PostDetail, error)
	FindByIDIgnoreActive(ctx context.Context, id uuid.UUID) (*model.Post, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) (*model.Post, error)
	UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]any) (*model.Post, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	IsInactiveByID(ctx context.Context, id uuid.UUID) (bool, error)
}
