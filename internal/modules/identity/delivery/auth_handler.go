package delivery

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authservice   service.AuthService
	githubservice service.GitHubOAuthService
}

func NewAuthHandler(
	authservice service.AuthService,
	githubservice service.GitHubOAuthService,
) *AuthHandler {
	return &AuthHandler{
		authservice:   authservice,
		githubservice: githubservice,
	}
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("login")

	req, err := httpx.Bind[dto.LoginRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			log.Warn("Login validation failed", slog.Any("validation_errors", validatorx.FormatValidationErrors(ve)))
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		log.Error("Invalid request body during login", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	log.Info("Login attempt", slog.String("email", req.Email))

	tokens, err := ah.authservice.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			log.Warn("Login failed - invalid credentials", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Email or password is incorrect")
		case service.ErrUserExistsWithGitHubLogin:
			log.Warn("Login failed - user exists with GitHub", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusConflict, httpx.ErrorCodeConflict, "This email is registered with GitHub. Please use GitHub login.")
		default:
			log.Error("Login failed - internal error", slog.String("email", req.Email), slog.Any("error", err))
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal server error")
		}
		return
	}

	log.Info("Login successful", slog.String("email", req.Email))

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (ah *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_login")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("GITHUB_CALLBACK_URL")

	if clientID == "" || redirectURI == "" {
		log.Error("GitHub OAuth not configured",
			slog.Bool("client_id_set", clientID != ""),
			slog.Bool("redirect_uri_set", redirectURI != ""))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "github oauth not configured")
		return
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("user:email"),
	)

	log.Info("Redirecting to GitHub OAuth", slog.String("auth_url", authURL))
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (ah *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_callback")

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Warn("GitHub callback missing code parameter")
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "missing code parameter")
		return
	}

	accessToken, err := ah.githubservice.ExchangeCodeForAccessToken(ctx, code)
	if err != nil {
		log.Error("Failed to exchange code for access token", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "error during github login")
		return
	}

	log.Debug("Access token obtained from GitHub")

	ghUser, err := ah.githubservice.GetUserInfo(ctx, accessToken)
	if err != nil {
		log.Error("Failed to get user info from GitHub", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to get user info")
		return
	}

	log.Info("GitHub user info retrieved",
		slog.String("github_email", ghUser.Email),
		slog.String("github_username", ghUser.Login),
		slog.Int64("github_id", ghUser.ID))

	tokens, err := ah.authservice.LoginWithGitHub(ctx, ghUser)
	if err != nil {
		switch err {
		case service.ErrUserExistsWithSystemLogin:
			log.Warn("GitHub login failed - user exists with system login", slog.String("email", ghUser.Email))
			httpx.WriteError(w, http.StatusConflict, httpx.ErrorCodeConflict, "This email is already registered. Please use email/password login instead.")
		case service.ErrUserExistsWithGitHubLogin:
			log.Warn("GitHub login failed - user already exists with GitHub", slog.String("email", ghUser.Email))
			httpx.WriteError(w, http.StatusConflict, httpx.ErrorCodeConflict, "This email is already registered. Please use github login instead.")
		default:
			log.Error("GitHub login failed",
				slog.String("email", ghUser.Email),
				slog.Int64("github_id", ghUser.ID),
				slog.Any("error", err))
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to login with github")
		}
		return
	}

	log.Info("GitHub login successful",
		slog.String("email", ghUser.Email),
		slog.Int64("github_id", ghUser.ID))

	httpx.WriteJSON(w, http.StatusOK, tokens)
}

func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httpx.Bind[dto.RefreshTokenRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	err = ah.authservice.Logout(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidRefreshToken:
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Invalid refresh token")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusNoContent, "")
}

func (ah *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httpx.Bind[dto.RefreshTokenRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	newAccessToken, newRefreshToken, err := ah.authservice.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken),
			errors.Is(err, service.ErrRefreshTokenExpired),
			errors.Is(err, service.ErrRefreshTokenRevoked):
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Invalid or expired refresh token")

		case errors.Is(err, service.ErrUserNotFound):
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "User not found")

		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"newAccessToken":     newAccessToken,
		"newRawRefreshToken": newRefreshToken,
	})
}
