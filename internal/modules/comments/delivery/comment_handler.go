package delivery

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/comments/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/comments/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CommentHandler struct {
	service service.CommentService
}

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (ch *CommentHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_comment")

	req, err := httpx.Bind[dto.CreateCommentRequest](r)
	if err != nil {
		log.Warn("Failed to bind request", slog.Any("error", err))
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	comment, err := req.ToModel()
	if err != nil {
		log.Warn("Failed to convert request to model", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid input data")
		return
	}

	createdComment, err := ch.service.CreateComment(ctx, &comment)
	if err != nil {
		log.Error("Failed to create comment", slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create comment")
		return
	}

	log.Info("Comment created", slog.String("id", createdComment.ID.String()), slog.String("post_id", createdComment.PostID.String()))
	res := dto.ToCommentFullResponse(createdComment)
	httpx.WriteJSON(w, 201, res)
}

func (ch *CommentHandler) ListPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_post_comments")

	postIDStr := r.PathValue("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		log.Warn("Invalid post ID in route param", slog.String("post_id", postIDStr), slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid post ID")
		return
	}

	comments, err := ch.service.ListPostComments(ctx, postID)
	if errors.Is(err, service.ErrPostNotFound) {
		log.Warn("Post not found for listing comments", slog.String("post_id", postID.String()))
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "Post not found")
		return
	}
	if errors.Is(err, service.ErrPostNotPublished) {
		log.Warn("Post not published for listing comments", slog.String("post_id", postID.String()))
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrorCodeForbidden, "Post is not published")
		return
	}
	if err != nil {
		log.Error("Failed to list comments", slog.String("post_id", postID.String()), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to list comments")
		return
	}

	log.Info("Comments listed successfully", slog.String("post_id", postID.String()), slog.Int("count", len(comments)))
	httpx.WriteJSON(w, 200, comments)
}
