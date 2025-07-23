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
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
	ErrUserNotFound        = errors.New("user not found")
)

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GitHubLoginResponse struct {
	Tokens TokensResponse `json:"tokens"`
	User   *model.User    `json:"user"`
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

func (as *AuthService) Login(ctx context.Context, email, password string) (*TokensResponse, error) {
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

	return as.generateTokensForUser(ctx, user)
}

func (as *AuthService) LoginWithGitHub(ctx context.Context, ghUser *model.GitHubUser) (*TokensResponse, error) {
	user, err := as.userRepo.FindByEmail(ctx, ghUser.Email)
	if errors.Is(err, sql.ErrNoRows) {
		// Create user if not exists
		user, err = as.createUserFromGitHub(ctx, ghUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user from github: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return as.generateTokensForUser(ctx, user)
}

func (as *AuthService) RefreshToken(ctx context.Context, refreshTokenInput string) (string, string, error) {
	hashed := jwtx.HashRefreshToken(refreshTokenInput)

	refreshToken, err := as.refreshTokenRepo.FindByHash(ctx, hashed)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return "", "", ErrRefreshTokenExpired
	}
	if refreshToken.RevokedAt != nil {
		return "", "", ErrRefreshTokenRevoked
	}

	if err := as.refreshTokenRepo.RevokeByID(ctx, refreshToken.ID); err != nil {
		return "", "", fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	user, err := as.userRepo.FindByID(ctx, refreshToken.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrUserNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to find user: %w", err)
	}

	newAccessToken, err := jwtx.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefreshToken, hashedRefreshToken, err := jwtx.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newRefreshToken := &model.RefreshToken{
		UserID:    refreshToken.UserID,
		TokenHash: hashedRefreshToken,
		UserAgent: logger.GetUserAgentFromContext(ctx),
		IPAddress: logger.GetIPAddressFromContext(ctx),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := as.refreshTokenRepo.Save(ctx, newRefreshToken); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return newAccessToken, rawRefreshToken, nil
}

func (as *AuthService) Logout(ctx context.Context, refreshTokenInput string) error {
	hashed := jwtx.HashRefreshToken(refreshTokenInput)

	refreshToken, err := as.refreshTokenRepo.FindByHash(ctx, hashed)
	if err != nil {
		return ErrInvalidRefreshToken
	}

	err = as.refreshTokenRepo.RevokeByID(ctx, refreshToken.ID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token by id: %v", err)
	}

	return nil
}

func (as *AuthService) CleanupExpiredOrRevokedTokens(ctx context.Context) error {
	return as.refreshTokenRepo.DeleteExpiredOrRevoked(ctx)
}

func (as *AuthService) generateTokensForUser(ctx context.Context, user *model.User) (*TokensResponse, error) {
	accessToken, err := jwtx.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, hashedRefreshToken, err := jwtx.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refresh := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashedRefreshToken,
		UserAgent: logger.GetUserAgentFromContext(ctx),
		IPAddress: logger.GetIPAddressFromContext(ctx),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := as.refreshTokenRepo.Save(ctx, refresh); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) createUserFromGitHub(ctx context.Context, ghUser *model.GitHubUser) (*model.User, error) {
	userFromGh := model.User{
		Name:  ghUser.Name,
		Email: ghUser.Email,
	}

	user, err := as.userRepo.Create(ctx, userFromGh)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
