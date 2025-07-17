package contracts

import (
	"context"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
)

type IUserRepository interface {
	Create(ctx context.Context, user model.User) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
