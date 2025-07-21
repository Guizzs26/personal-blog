package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts"
	"github.com/Guizzs26/personal-blog/pkg/hashx"
	"github.com/Guizzs26/personal-blog/pkg/jwtx"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type AuthService struct {
	repo contracts.IUserRepository
}

func NewAuthService(repo contracts.IUserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (as *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := as.repo.FindByEmail(ctx, email)

	// try to prevent timing attack
	validPassword := false
	if err == nil {
		validPassword = hashx.Compare(user.Password, password)
	} else {
		hashx.Compare("dummyPassword", password)
	}

	if errors.Is(err, sql.ErrNoRows) || !validPassword {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("an error occurred when searching for a user by email: %v", err)
	}

	accessToken, err := jwtx.GenerateToken(user.ID.String(), email)
	if err != nil {
		return "", fmt.Errorf("an error occurred when generating access jwt token: %v", err)
	}

	return accessToken, nil
}
