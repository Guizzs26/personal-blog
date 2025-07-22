package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/Guizzs26/personal-blog/pkg/hashx"
	"github.com/Guizzs26/personal-blog/pkg/jwtx"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type TokensRes struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	userRepo         contracts.IUserRepository
	refreshTokenRepo contracts.IRefreshTokenRepository
}

func NewAuthService(
	userRepo contracts.IUserRepository,
	refreshTokenRepo contracts.IRefreshTokenRepository,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (as *AuthService) Login(ctx context.Context, email, password string) (*TokensRes, error) {
	user, err := as.userRepo.FindByEmail(ctx, email)

	// try to prevent timing attack
	validPassword := false
	if err == nil {
		validPassword = hashx.Compare(user.Password, password)
	} else {
		hashx.Compare("dummyPassword", password)
	}

	if errors.Is(err, sql.ErrNoRows) || !validPassword {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("an error occurred when searching for a user by email: %v", err)
	}

	accessToken, err := jwtx.GenerateAccessToken(user.ID.String(), email)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when generating access jwt token: %v", err)
	}

	refreshToken, hashedRefreshToken, err := jwtx.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("an error occurred when generating refresh token: %v", err)
	}

	refresh := model.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefreshToken,
		UserAgent: logger.GetUserAgentFromContext(ctx),
		IPAddress: logger.GetIPAddressFromContext(ctx),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = as.refreshTokenRepo.Save(ctx, &refresh)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when saving refresh token in database: %v", err)
	}

	return &TokensRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
