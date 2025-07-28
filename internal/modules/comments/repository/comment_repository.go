package repository

import (
	"context"
	"database/sql"
	"errors"
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
            status, active, is_pinned, created_at, updated_at
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
            status, active, is_pinned, created_at, updated_at
        FROM comments 
        WHERE id = $1
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find comment by id: %v", err), 0)
	}

	return &comment, nil
}

func (pcr *PostgresCommentsRepository) FindAllByPostID(ctx context.Context, postID uuid.UUID) ([]model.Comment, error) {
	query := `
			WITH ordered_comments AS (
				SELECT
					id, post_id, user_id, parent_comment_id, content,
					status, is_pinned, active, created_at, updated_at,
					CASE 	
						WHEN parent_comment_id IS NULL THEN 0
					END as comment_level
				FROM comments
				WHERE post_id = $1
					AND active = true
			)
			SELECT id, post_id, user_id, parent_comment_id, content,
				status, is_pinned, active, created_at, updated_at
			FROM ordered_comments
			ORDER BY
				comment_level ASC,
				is_pinned DESC,
				created_at ASC
			`

	rows, err := pcr.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find comments by post id: %v", err), 0)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentCommentID,
			&comment.Content,
			&comment.Status,
			&comment.IsPinned,
			&comment.Active,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan comment: %v", err), 0)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("error iterating comments: %v", err), 0)
	}

	return comments, nil
}

func (pcr *PostgresCommentsRepository) FindByIDIgnoreActive(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	query := `
        SELECT 
            id, post_id, user_id, parent_comment_id, content,
            status, active, is_pinned, created_at, updated_at
        FROM comments 
        WHERE id = $1
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find comment by id: %v", err), 0)
	}

	return &comment, nil
}

func (pcr *PostgresCommentsRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) (*model.Comment, error) {
	query := `
        UPDATE comments 
        SET active = $1, 
                updated_at = NOW() 
        WHERE id = $2
        RETURNING id, post_id, user_id, parent_comment_id, content,
            status, active, is_pinned, created_at, updated_at
    `

	row := pcr.db.QueryRowContext(ctx, query, active, id)
	var comment model.Comment
	err := row.Scan(
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
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to set comment active status: %v", err), 0)
	}

	return &comment, nil
}

func (pcr *PostgresCommentsRepository) SetPinned(ctx context.Context, id uuid.UUID, isPinned bool) (*model.Comment, error) {
	query := `
        UPDATE comments 
        SET is_pinned = $1, 
                updated_at = NOW() 
        WHERE id = $2
        RETURNING id, post_id, user_id, parent_comment_id, content,
            status, active, is_pinned, created_at, updated_at
    `

	row := pcr.db.QueryRowContext(ctx, query, isPinned, id)
	var comment model.Comment
	err := row.Scan(
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
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to set comment pinned status: %v", err), 0)
	}

	return &comment, nil
}

func (pcr *PostgresCommentsRepository) FindPendingForModeration(ctx context.Context) ([]model.Comment, error) {
	query := `
        SELECT 
            id, post_id, user_id, parent_comment_id, content,
            status, active, is_pinned, created_at, updated_at
        FROM comments 
        WHERE status = 'pending' AND active = true
        ORDER BY created_at ASC
    `

	rows, err := pcr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to find pending comments for moderation: %v", err), 0)
	}
	defer rows.Close()

	var comments []model.Comment
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
		)
		if err != nil {
			return nil, xerrors.WithStackTrace(fmt.Errorf("failed to scan pending comment: %v", err), 0)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("error iterating pending comments: %v", err), 0)
	}

	return comments, nil
}

func (pcr *PostgresCommentsRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `
        DELETE FROM comments 
        WHERE id = $1
    `

	_, err := pcr.db.ExecContext(ctx, query, id)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("failed to delete comment: %v", err), 0)
	}

	return nil
}

func (pcr *PostgresCommentsRepository) UpdateByID(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	query := `
        UPDATE comments
        SET content = $1,
            updated_at = NOW()
        WHERE id = $2
        RETURNING id, post_id, user_id, parent_comment_id, content,
                  status, active, is_pinned, created_at, updated_at
    `

	row := pcr.db.QueryRowContext(ctx, query, comment.Content, comment.ID)
	var updated model.Comment
	err := row.Scan(
		&updated.ID,
		&updated.PostID,
		&updated.UserID,
		&updated.ParentCommentID,
		&updated.Content,
		&updated.Status,
		&updated.Active,
		&updated.IsPinned,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("failed to update comment: %v", err), 0)
	}

	return &updated, nil
}
