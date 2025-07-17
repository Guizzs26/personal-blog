package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	categoryInterfaces "github.com/Guizzs26/personal-blog/internal/modules/categories/contracts/interfaces"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/interfaces"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
	"github.com/Guizzs26/personal-blog/pkg/slug"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrPostIsActive = errors.New("post inactive")
)

// PostService contains business logic for managing posts
type PostService struct {
	repo         interfaces.IPostRepository
	categoryRepo categoryInterfaces.ICategoryRepository
}

// NewPostService creates a new PostService with the given repository
func NewPostService(repo interfaces.IPostRepository, categoryRepo categoryInterfaces.ICategoryRepository) *PostService {
	return &PostService{
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

// CreatePost creates a new post, generating a unique slug based on its title
func (ps *PostService) CreatePost(ctx context.Context, post model.Post) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_post_service")

	if post.Published {
		if post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
	} else {
		post.PublishedAt = nil // just a guarantee
	}

	slug, err := ps.generateUniqueSlug(ctx, post.Title)
	if err != nil {
		log.Error("Failed to generate unique slug", slog.String("title", post.Title), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to generate unique slug"), err)
	}

	existsCategory, err := ps.categoryRepo.ExistsByID(ctx, post.CategoryID)
	if err != nil {
		log.Error("Failed to validate category existence",
			slog.String("category_id", post.CategoryID.String()), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to validate category existence"), err)
	}

	if !existsCategory {
		log.Warn("Category does not exist", slog.String("category_id", post.CategoryID.String()))
		return nil, xerrors.New("informed category does not exist")
	}

	post.Slug = slug
	createdPost, err := ps.repo.Create(ctx, post)
	if err != nil {
		log.Error("Failed to create post", slog.String("slug", post.Slug), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create post"), err)
	}

	log.Info("Post created", slog.String("id", createdPost.ID.String()), slog.String("slug", createdPost.Slug))
	return createdPost, nil
}

func (ps *PostService) ListPublishedAndPaginatedPosts(ctx context.Context, page, pageSize int, categorySlug *string) ([]model.PostPreview, int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_published_posts_service")

	var posts []model.PostPreview
	var totalCount int
	var postsErr, countErr error

	done := make(chan bool, 2)

	go func() {
		posts, postsErr = ps.repo.ListPublished(ctx, page, pageSize, categorySlug)
		done <- true
	}()

	go func() {
		totalCount, countErr = ps.repo.CountPublished(ctx, categorySlug)
		done <- true
	}()

	<-done
	<-done

	if postsErr != nil {
		log.Error("Failed to list published posts", slog.Any("error", postsErr))
		return nil, 0, xerrors.WithWrapper(xerrors.New("failed to list published posts"), postsErr)
	}

	if countErr != nil {
		log.Error("Failed to count published posts", slog.Any("error", countErr))
		return nil, 0, xerrors.WithWrapper(xerrors.New("failed to count published posts"), countErr)
	}

	return posts, totalCount, nil
}

func (ps *PostService) GetPublishedPostBySlug(ctx context.Context, slug string) (*model.PostDetail, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("get_published_post_service")

	post, err := ps.repo.FindPublishedBySlug(ctx, slug)
	if errors.Is(err, repository.ErrResourceNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		log.Error("Failed to find post by slug", slog.String("slug", slug), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to find post by slug"), err)
	}

	return post, nil
}

func (ps *PostService) SetPostActive(ctx context.Context, id uuid.UUID, active bool) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("set_active_post_service")

	existingPost, err := ps.repo.FindByIDIgnoreActive(ctx, id)
	if errors.Is(err, repository.ErrResourceNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		log.Error("Failed to find post", slog.String("id", id.String()), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to find post"), err)
	}
	if existingPost.Active == active {
		return existingPost, nil
	}

	post, err := ps.repo.SetActive(ctx, id, active)
	if err != nil {
		log.Error("Failed to update post status", slog.String("id", id.String()), slog.Bool("active", active), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to update post status"), err)
	}

	log.Info("Post status updated", slog.String("id", post.ID.String()), slog.String("slug", post.Slug), slog.Bool("active", active))
	return post, nil
}

func (ps *PostService) UpdatePostByID(ctx context.Context, id uuid.UUID, updates map[string]any) (*model.Post, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("post_service")

	if newTitleRaw, ok := updates["title"]; ok {
		if newTitle, ok := newTitleRaw.(string); ok && strings.TrimSpace(newTitle) != "" {
			slug, err := ps.generateUniqueSlug(ctx, newTitle)
			if err != nil {
				return nil, xerrors.WithWrapper(xerrors.New("failed to generate slug"), err)
			}
			log.Debug("Generating new slug for updated title", slog.String("new_title", newTitle), slog.String("new_slug", slug))
			updates["slug"] = slug
		}
	}

	if publishedRaw, ok := updates["published"]; ok {
		if published, ok := publishedRaw.(bool); ok {
			if published {
				log.Debug("Setting post as published", slog.Time("published_at", time.Now()))
				updates["published_at"] = time.Now()
			} else {
				log.Debug("Setting post as unpublished")
				updates["published_at"] = nil
			}
		}
	}

	updatedPost, err := ps.repo.UpdateByID(ctx, id, updates)
	if errors.Is(err, repository.ErrResourceNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		log.Error("Failed to update post", slog.String("id", id.String()), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to update post"), err)
	}

	log.Info("Post updated", slog.String("id", updatedPost.ID.String()))
	return updatedPost, nil
}

func (ps *PostService) DeletePostByID(ctx context.Context, id uuid.UUID) error {
	log := logger.GetLoggerFromContext(ctx).WithGroup("post_service")

	isInactive, err := ps.repo.IsInactiveByID(ctx, id)
	if errors.Is(err, repository.ErrResourceNotFound) {
		log.Warn("Post not found for deletion", slog.String("id", id.String()))
		return ErrPostNotFound
	}
	if err != nil {
		log.Error("Failed to check if post is inactive", slog.String("id", id.String()), slog.Any("error", err))
		return xerrors.WithWrapper(xerrors.New("service: check inactive by id"), err)
	}
	if !isInactive {
		log.Warn("Cannot delete post because it is still active", slog.String("id", id.String()))
		return ErrPostIsActive
	}

	err = ps.repo.DeleteByID(ctx, id)
	if errors.Is(err, repository.ErrResourceNotFound) {
		log.Warn("Post not found for deletion", slog.String("id", id.String()))
		return ErrPostNotFound
	}
	if err != nil {
		log.Error("Failed to delete post", slog.String("id", id.String()), slog.Any("error", err))
		return xerrors.WithWrapper(xerrors.New("service: delete post by id"), err)
	}

	log.Debug("Post deleted successfully", slog.String("id", id.String()))
	return nil
}

// generateUniqueSlug ensures that the generated slug is unique in the database
func (ps *PostService) generateUniqueSlug(ctx context.Context, t string) (string, error) {
	log := logger.GetLoggerFromContext(ctx)

	baseSlug := slug.GenerateSlug(t)
	slug := baseSlug

	exists, err := ps.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		log.Error("Failed to check slug existence",
			slog.String("slug", slug),
			slog.Any("error", err))

		return "", xerrors.WithWrapper(xerrors.New("service: check if slug exists"), err)
	}

	if !exists {
		return slug, nil
	}

	// Slug already exists, try variations
	for i := 1; ; i++ {
		slug = fmt.Sprintf("%s-%d", baseSlug, i)

		exists, err := ps.repo.ExistsBySlug(ctx, slug)
		if err != nil {
			log.Error("Failed to check slug existence in loop",
				slog.String("slug", slug),
				slog.Int("attempt", i),
				slog.Any("error", err))

			return "", xerrors.WithWrapper(xerrors.New(fmt.Sprintf("service: check slug existence in variation (attempt %d)", i)), err)
		}

		if !exists {
			break
		}
	}
	return slug, nil
}
