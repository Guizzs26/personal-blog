package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/Guizzs26/personal-blog/pkg/hashx"
	"github.com/Guizzs26/personal-blog/pkg/jwtx"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrInvalidRefreshToken       = errors.New("invalid refresh token")
	ErrRefreshTokenExpired       = errors.New("refresh token expired")
	ErrRefreshTokenRevoked       = errors.New("refresh token revoked")
	ErrUserNotFound              = errors.New("user not found")
	ErrUserExistsWithSystemLogin = errors.New("user already exists with system login")
	ErrUserExistsWithGitHubLogin = errors.New("user already exists with github login")
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
	log := logger.GetLoggerFromContext(ctx).WithGroup("login_service")

	log.Debug("Starting login process", slog.String("email", email))

	user, err := as.userRepo.FindByEmail(ctx, email)

	// try to prevent timing attack
	validPassword := false
	if err == nil {
		log.Debug("User found in database", slog.String("user_id", user.ID.String()))

		// Check if user was created via GitHub (has github_id)
		if user.GitHubID != nil {
			log.Warn("Login attempt for GitHub user with system credentials",
				slog.String("email", email),
				slog.Int64("github_id", *user.GitHubID))
			// User was created through GitHub, prohibit system login
			hashx.Compare("dummyPassword", password) // prevent timing attack
			return nil, ErrUserExistsWithGitHubLogin
		}
		validPassword = hashx.Compare(user.Password, password)
		log.Debug("Password validation completed", slog.Bool("valid", validPassword))
	} else {
		log.Debug("User not found, running dummy hash comparison", slog.String("email", email))
		hashx.Compare("dummyPassword", password)
	}

	if errors.Is(err, sql.ErrNoRows) || !validPassword {
		log.Warn("Login failed - invalid credentials", slog.String("email", email))
		return nil, ErrUserNotFound
	}
	if err != nil {
		log.Error("Database error during login", slog.String("email", email), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to find user by email"), err)
	}

	log.Info("Login successful, generating tokens", slog.String("user_id", user.ID.String()))
	return as.generateTokensForUser(ctx, user)
}

func (as *AuthService) LoginWithGitHub(ctx context.Context, ghUser *model.GitHubUser) (*TokensResponse, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_login_service")

	log.Info("Starting GitHub login process",
		slog.String("github_email", ghUser.Email),
		slog.Int64("github_id", ghUser.ID),
		slog.String("github_username", ghUser.Login))

	// First attempt: search by GitHub ID (more reliable)
	user, err := as.userRepo.FindByGitHubID(ctx, ghUser.ID)
	if err == nil {
		log.Info("User found by GitHub ID",
			slog.String("user_id", user.ID.String()),
			slog.String("current_email", user.Email),
			slog.String("github_email", ghUser.Email))

		// Check if the email has changed and update if necessary
		if user.Email != ghUser.Email {
			log.Info("Updating user email from GitHub",
				slog.String("user_id", user.ID.String()),
				slog.String("old_email", user.Email),
				slog.String("new_email", ghUser.Email))

			user.Email = ghUser.Email
			if err := as.userRepo.Update(ctx, user); err != nil {
				log.Error("Failed to update user email",
					slog.String("user_id", user.ID.String()),
					slog.String("new_email", ghUser.Email),
					slog.Any("error", err))
				return nil, xerrors.WithWrapper(xerrors.New("failed to update user email"), err)
			}
			log.Info("User email updated successfully", slog.String("user_id", user.ID.String()))
		}

		log.Info("GitHub login successful for existing user", slog.String("user_id", user.ID.String()))
		return as.generateTokensForUser(ctx, user)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Error("Database error while finding user by GitHub ID",
			slog.Int64("github_id", ghUser.ID),
			slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to find user by github id"), err)
	}

	log.Debug("User not found by GitHub ID, checking by email", slog.String("email", ghUser.Email))

	// Second attempt: search by email to check if user already exists
	_, err = as.userRepo.FindByEmail(ctx, ghUser.Email)
	if err == nil {
		log.Warn("GitHub login blocked - user exists with system login",
			slog.String("email", ghUser.Email),
			slog.Int64("github_id", ghUser.ID))
		// User exists but was created through the system (not GitHub)
		// We prohibit GitHub login for system-created users
		return nil, ErrUserExistsWithSystemLogin
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Error("Database error while finding user by email",
			slog.String("email", ghUser.Email),
			slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to find user by email"), err)
	}

	log.Info("Creating new user from GitHub",
		slog.String("email", ghUser.Email),
		slog.Int64("github_id", ghUser.ID),
		slog.String("name", ghUser.Name))

	// User does not exist - create new (only GitHub users from now on)
	user, err = as.createUserFromGitHub(ctx, ghUser)
	if err != nil {
		log.Error("Failed to create user from GitHub",
			slog.String("email", ghUser.Email),
			slog.Int64("github_id", ghUser.ID),
			slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create user from github"), err)
	}

	log.Info("New user created from GitHub",
		slog.String("user_id", user.ID.String()),
		slog.String("email", user.Email))

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
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_user_from_github")

	user := model.User{
		Name:     ghUser.Name,
		Email:    ghUser.Email,
		GitHubID: &ghUser.ID,
		Active:   true,
	}

	log.Debug("Creating user from GitHub data",
		slog.String("name", user.Name),
		slog.String("email", user.Email),
		slog.Int64("github_id", *user.GitHubID))

	createdUser, err := as.userRepo.CreateFromGitHub(ctx, user)
	if err != nil {
		log.Error("Failed to create user in database",
			slog.String("email", user.Email),
			slog.Int64("github_id", *user.GitHubID),
			slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create user"), err)
	}

	log.Info("User created successfully from GitHub",
		slog.String("user_id", createdUser.ID.String()),
		slog.String("email", createdUser.Email))

	return createdUser, nil
}
