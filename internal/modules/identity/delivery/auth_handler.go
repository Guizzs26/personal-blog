package delivery

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

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

	req, err := httpx.Bind[dto.LoginRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	tokens, err := ah.authservice.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Email or password is incorrect")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Internal server error")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (ah *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURI := os.Getenv("GITHUB_CALLBACK_URL")

	if clientID == "" || redirectURI == "" {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "github oauth not configured")
		return
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape("user:email"),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (ah *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "missing code parameter")
		return
	}

	accessToken, err := ah.githubservice.ExchangeCodeForAccessToken(code)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "error during github login")
		return
	}

	ghUser, err := ah.githubservice.GetUserInfo(accessToken)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to get user info")
		return
	}

	tokens, err := ah.authservice.LoginWithGitHub(ctx, ghUser)
	if err != nil {
		switch err {
		case service.ErrUserExistsWithSystemLogin:
			httpx.WriteError(w, http.StatusConflict, httpx.ErrorCodeConflict, "This email is already registered. Please use email/password login instead.")
		case service.ErrUserExistsWithGitHubLogin:
			httpx.WriteError(w, http.StatusConflict, httpx.ErrorCodeConflict, "This email is already registered. Please use github login instead.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to login with github")
		}
		return
	}

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
