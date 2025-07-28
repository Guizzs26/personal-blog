package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
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
	log := logger.GetLoggerFromContext(ctx).WithGroup("comment_service")

	if comment.ParentCommentID != nil {
		log.Debug("Checking parent comment existence",
			slog.String("parent_comment_id", comment.ParentCommentID.String()),
			slog.String("post_id", comment.PostID.String()),
		)
		existing, err := cs.repo.FindByID(ctx, *comment.ParentCommentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Warn("Parent comment not found",
					slog.String("parent_comment_id", comment.ParentCommentID.String()),
					slog.String("post_id", comment.PostID.String()),
					slog.Any("error", err),
				)
				return nil, xerrors.WithWrapper(xerrors.New("parent comment not found"), err)
			}
			log.Error("Error when checking parent comment",
				slog.String("parent_comment_id", comment.ParentCommentID.String()),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("error when checking parent comment: %v", err)
		}

		if existing.PostID != comment.PostID {
			log.Warn("Parent comment does not belong to the same post",
				slog.String("parent_comment_id", comment.ParentCommentID.String()),
				slog.String("post_id", comment.PostID.String()),
			)
			return nil, xerrors.New("parent comment does not belong to the same post")
		}

		if existing.ParentCommentID != nil {
			log.Warn("Replies to replies are not allowed",
				slog.String("parent_comment_id", comment.ParentCommentID.String()),
				slog.String("post_id", comment.PostID.String()),
			)
			return nil, xerrors.New("replies to replies are not allowed")
		}
	}

	createdComment, err := cs.repo.Create(ctx, comment)
	if err != nil {
		log.Error("Failed to create comment",
			slog.String("post_id", comment.PostID.String()),
			slog.Any("error", err),
		)
		return nil, xerrors.WithWrapper(xerrors.New("failed to create comment"), err)
	}

	log.Info("Comment created successfully",
		slog.String("comment_id", createdComment.ID.String()),
		slog.String("post_id", createdComment.PostID.String()),
	)
	return createdComment, nil
}

func (cs *CommentService) ListPostComments(ctx context.Context, postID uuid.UUID) ([]CommentResponse, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("comment_service")

	log.Debug("Checking post existence for comments", slog.String("post_id", postID.String()))
	post, err := cs.postRepo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn("Post not found for listing comments", slog.String("post_id", postID.String()))
			return nil, ErrPostNotFound
		}
		log.Error("Error when checking post existence", slog.String("post_id", postID.String()), slog.Any("error", err))
		return nil, fmt.Errorf("error when checking post existence: %v", err)
	}

	if !post.Published {
		log.Warn("Post not published for listing comments", slog.String("post_id", postID.String()))
		return nil, ErrPostNotPublished
	}

	log.Debug("Listing comments for post", slog.String("post_id", postID.String()))
	comments, err := cs.repo.FindAllByPostID(ctx, postID)
	if err != nil {
		log.Error("Failed to list comments", slog.String("post_id", postID.String()), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to list comments"), err)
	}

	log.Info("Comments listed successfully",
		slog.String("post_id", postID.String()),
		slog.Int("count", len(comments)),
	)
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
