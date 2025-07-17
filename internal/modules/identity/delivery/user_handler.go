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

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (uh *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httpx.Bind[dto.CreateUserRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	user := req.ToModel()

	createdUser, err := uh.service.CreateUser(ctx, user)
	if errors.Is(err, service.ErrEmailAlreadyInUse) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "email already in use")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "somethin went wrong when creating user")
		return
	}

	res := dto.ToUserFullResponse(createdUser)
	httpx.WriteJSON(w, 201, res)
}
