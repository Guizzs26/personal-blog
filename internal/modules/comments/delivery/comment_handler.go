package delivery

import (
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/comments/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
)

type CommentHandler struct {
	service service.CommentService
}

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (ch *CommentHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httpx.Bind[dto.CreateCommentRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	comment, err := req.ToModel()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid input data")
		return
	}

	createdComment, err := ch.service.CreateComment(ctx, &comment)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create comment")
		return
	}

	res := dto.ToCommentFullResponse(createdComment)
	httpx.WriteJSON(w, 201, res)
}
