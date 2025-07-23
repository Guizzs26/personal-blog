package contracts

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/google/uuid"
)

type IUserRepository interface {
	Create(ctx context.Context, user model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByGitHubID(ctx context.Context, gitHubID int64) (*model.User, error)
}

type IRefreshTokenRepository interface {
	Save(ctx context.Context, token *model.RefreshToken) error
	RevokeByID(ctx context.Context, id uuid.UUID) error
	DeleteExpiredOrRevoked(ctx context.Context) error
	FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
}
