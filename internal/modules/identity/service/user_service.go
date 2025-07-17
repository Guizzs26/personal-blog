package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/Guizzs26/personal-blog/pkg/hashx"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrEmailAlreadyInUse = errors.New("email already taken")
)

type UserService struct {
	repo contracts.IUserRepository
}

func NewUserService(repo contracts.IUserRepository) *UserService {
	return &UserService{repo: repo}
}

func (us *UserService) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	existingUser, err := us.repo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return nil, xerrors.WithWrapper(xerrors.New("error verifying user"), err)
	}

	if existingUser {
		return nil, fmt.Errorf("user email already in use: %v", ErrEmailAlreadyInUse)
	}

	hashedPass, err := hashx.Generate(user.Password)
	if err != nil {
		return nil, fmt.Errorf("create-user: error hash password: %v", err)
	}

	userToCreate := user
	userToCreate.Password = hashedPass

	createdUser, err := us.repo.Create(ctx, userToCreate)
	if err != nil {
		return nil, xerrors.WithWrapper(xerrors.New("failed to create user"), err)
	}

	createdUser.Password = ""

	return createdUser, nil
}
