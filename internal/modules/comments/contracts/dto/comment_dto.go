package dto

import (
	"fmt"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	"github.com/google/uuid"
)

type CreateCommentRequest struct {
	PostID          string  `json:"post_id"  validate:"required,uuid4"`
	UserID          string  `json:"user_id" validate:"required,uuid4"`
	ParentCommentID *string `json:"parent_comment_id,omitempty" validate:"omitempty,uuid4"`
	Content         string  `json:"content" validate:"required,min=1,max=500"`
}

func (ccr *CreateCommentRequest) ToModel() (model.Comment, error) {
	postUUID, err := uuid.Parse(ccr.PostID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("failed to parse post_id to a valid uuid: %w", err)
	}

	userUUID, err := uuid.Parse(ccr.UserID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("failed to parse user_id to a valid uuid: %w", err)
	}

	var parentCommentUUID *uuid.UUID
	if ccr.ParentCommentID != nil {
		parsedParentCommentID, err := uuid.Parse(*ccr.ParentCommentID)
		if err != nil {
			return model.Comment{}, fmt.Errorf("failed to parse parent_comment_id to a valid uuid: %w", err)
		}
		parentCommentUUID = &parsedParentCommentID
	}

	return model.Comment{
		PostID:          postUUID,
		UserID:          userUUID,
		ParentCommentID: parentCommentUUID,
		Content:         ccr.Content,
	}, nil
}

type CommentFullResponse struct {
	ID              string    `json:"id"`
	PostID          string    `json:"post_id"`
	UserID          string    `json:"user_id"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	Content         string    `json:"content"`
	Status          string    `json:"status"`
	Active          bool      `json:"active"`
	IsPinned        bool      `json:"is_pinned"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func ToCommentFullResponse(comment *model.Comment) *CommentFullResponse {
	var parentCommentID *string
	if comment.ParentCommentID != nil {
		id := comment.ParentCommentID.String()
		parentCommentID = &id
	}

	return &CommentFullResponse{
		ID:              comment.ID.String(),
		PostID:          comment.PostID.String(),
		UserID:          comment.UserID.String(),
		ParentCommentID: parentCommentID,
		Content:         comment.Content,
		Status:          comment.Status,
		Active:          comment.Active,
		IsPinned:        comment.IsPinned,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}
}

type ListPostCommentsRequest struct {
	PostID string `json:"post_id" validate:"required,uuid4"`
}

func (lpc *ListPostCommentsRequest) ToModel() (uuid.UUID, error) {
	postUUID, err := uuid.Parse(lpc.PostID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse post_id to a valid uuid: %w", err)
	}
	return postUUID, nil
}
