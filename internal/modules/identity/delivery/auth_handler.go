package delivery

import (
	"errors"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
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

	tokens, err := ah.service.Login(ctx, req.Email, req.Password)
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

	err = ah.service.Logout(ctx, req.RefreshToken)
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

	newAccessToken, newRefreshToken, err := ah.service.RefreshToken(ctx, req.RefreshToken)
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
