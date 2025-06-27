package interfaces

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
)

type IPostRepository interface {
	Create(ctx context.Context, post model.Post) (*model.Post, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ListPublished(ctx context.Context, page, pageSize int) ([]model.PostPreview, error)
	CountPublished(ctx context.Context) (int, error)
}
