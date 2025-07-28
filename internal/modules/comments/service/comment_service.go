package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/comments/contracts"
	"github.com/Guizzs26/personal-blog/internal/modules/comments/model"
	postContracts "github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/interfaces"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrPostNotPublished = errors.New("post is not available")
)

type CommentResponse struct {
	model.Comment
	Replies []CommentResponse `json:"replies,omitempty"`
}

type CommentService struct {
	repo     contracts.ICommentRepository
	postRepo postContracts.IPostRepository
}

func NewCommentService(
	repo contracts.ICommentRepository,
	postRepo postContracts.IPostRepository) *CommentService {
	return &CommentService{
		repo:     repo,
		postRepo: postRepo,
	}
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

func (cs *CommentService) ListPostComments(ctx context.Context, postID uuid.UUID) ([]CommentResponse, error) {
	post, err := cs.postRepo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error when checking post existence: %v", err)
	}

	if !post.Published {
		return nil, ErrPostNotPublished
	}

	comments, err := cs.repo.FindAllByPostID(ctx, postID)
	if err != nil {
		return nil, xerrors.WithWrapper(xerrors.New("failed to list comments"), err)
	}

	return cs.organizeCommentsHierarchy(comments), nil
}

func (cs *CommentService) organizeCommentsHierarchy(comments []model.Comment) []CommentResponse {
	commentMap := make(map[uuid.UUID]CommentResponse)
	var topLevelComments []CommentResponse

	// create map for all comments
	for _, comment := range comments {
		commentMap[comment.ID] = CommentResponse{
			Comment: comment,
			Replies: []CommentResponse{},
		}
	}

	// organize hierarchy
	for _, comment := range comments {
		if comment.ParentCommentID == nil {
			topLevelComments = append(topLevelComments, commentMap[comment.ID])
		} else {
			// Reply - add to parent's replies
			if parent, exists := commentMap[*comment.ParentCommentID]; exists {
				parent.Replies = append(parent.Replies, commentMap[comment.ID])
				commentMap[*comment.ParentCommentID] = parent
			}
		}
	}
	return topLevelComments
}
