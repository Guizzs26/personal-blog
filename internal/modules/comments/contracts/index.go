package contracst

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	"github.com/google/uuid"
)

type ICommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error)
	FindPendingForModeration(ctx context.Context) ([]*model.Comment, error)
}
