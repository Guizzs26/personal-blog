package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

type PostgresCommentsRepository struct {
	db *sql.DB
}

func NewPostgresCommentsRepository(db *sql.DB) *PostgresCommentsRepository {
	return &PostgresCommentsRepository{db: db}
}

func (pcr *PostgresCommentsRepository) Create(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	query := `
		INSERT INTO comments
			(post_id, user_id, parent_comment_id, content)
		VALUES 
			($1, $2, $3, $4)
		RETURNING
			id, post_id, user_id, parent_comment_id, content,
			status, active, is_pinned, created_at, updated_at, deleted_at
	`

	var createdComment model.Comment
	err := pcr.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.UserID,
		comment.ParentCommentID,
		comment.Content,
	).Scan(
		&createdComment.ID,
		&createdComment.PostID,
		&createdComment.UserID,
		&createdComment.ParentCommentID,
		&createdComment.Content,
		&createdComment.Status,
		&createdComment.Active,
		&createdComment.IsPinned,
		&createdComment.CreatedAt,
		&createdComment.UpdatedAt,
		&createdComment.DeletedAt,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to create comment: %v", err), 0)
	}

	return &createdComment, nil
}

func (pcr *PostgresCommentsRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	query := `
		SELECT 
			id, post_id, user_id, parent_comment_id, content,
			status, active, is_pinned, created_at, updated_at, deleted_at
		FROM comments 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var comment model.Comment
	err := pcr.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.Status,
		&comment.Active,
		&comment.IsPinned,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find comment by id: %v", err), 0)
	}

	return &comment, nil
}

func (pcr *PostgresCommentsRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error) {
	query := `
		SELECT 
			id, post_id, user_id, parent_comment_id, content,
			status, active, is_pinned, created_at, updated_at, deleted_at
		FROM comments 
		WHERE post_id = $1 AND deleted_at IS NULL AND active = true
		ORDER BY is_pinned DESC, created_at ASC
	`

	rows, err := pcr.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find comments by post id: %v", err), 0)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentCommentID,
			&comment.Content,
			&comment.Status,
			&comment.Active,
			&comment.IsPinned,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.DeletedAt,
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan comment: %v", err), 0)
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("error iterating comments: %v", err), 0)
	}

	return comments, nil
}

func (pcr *PostgresCommentsRepository) FindPendingForModeration(ctx context.Context) ([]*model.Comment, error) {
	query := `
		SELECT 
			id, post_id, user_id, parent_comment_id, content,
			status, active, is_pinned, created_at, updated_at, deleted_at
		FROM comments 
		WHERE status = 'pending' AND deleted_at IS NULL AND active = true
		ORDER BY created_at ASC
	`

	rows, err := pcr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find pending comments for moderation: %v", err), 0)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentCommentID,
			&comment.Content,
			&comment.Status,
			&comment.Active,
			&comment.IsPinned,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.DeletedAt,
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan pending comment: %v", err), 0)
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("error iterating pending comments: %v", err), 0)
	}

	return comments, nil
}
