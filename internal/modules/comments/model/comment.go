package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	PostID          uuid.UUID  `json:"post_id" db:"post_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	ParentCommentID *uuid.UUID `json:"parent_comment_id" db:"parent_comment_id"`
	Content         string     `json:"content" db:"content"`
	Status          string     `json:"status" db:"status"`
	Active          bool       `json:"active" db:"active"`
	IsPinned        bool       `json:"is_pinned" db:"is_pinned"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
