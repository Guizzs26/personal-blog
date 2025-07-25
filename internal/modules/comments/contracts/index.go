package contracst

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	"github.com/google/uuid"
)

type ICommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	FindAllByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error)
	FindByIDIgnoreActive(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) (*model.Comment, error)
	SetPinned(ctx context.Context, id uuid.UUID, isPinned bool) (*model.Comment, error)
	FindPendingForModeration(ctx context.Context) ([]model.Comment, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	UpdateByID(ctx context.Context, comment *model.Comment) (*model.Comment, error)
}
