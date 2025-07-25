package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	contracst "github.com/Guizzs26/personal-blog/internal/modules/comments/contracts"
	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	"github.com/mdobak/go-xerrors"
)

type CommentService struct {
	repo contracst.ICommentRepository
}

func NewCommentService(repo contracst.ICommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (cs *CommentService) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	if comment.ParentCommentID != nil {
		existing, err := cs.repo.FindByID(ctx, *comment.ParentCommentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, xerrors.WithWrapper(xerrors.New("parent comment not found"), err)
			}
			return nil, fmt.Errorf("error when checking parent comment: %v", err)
		}

		if existing.PostID != comment.PostID {
			return nil, xerrors.New("parent comment does not belong to the same post")
		}

		if existing.ParentCommentID != nil {
			return nil, xerrors.New("replies to replies are not allowed")
		}
	}

	createdComment, err := cs.repo.Create(ctx, comment)
	if err != nil {
		return nil, xerrors.WithWrapper(xerrors.New("failed to create comment"), err)
	}

	return createdComment, nil
}
