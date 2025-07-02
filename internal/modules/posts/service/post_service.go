package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/contracts/interfaces"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrPostIsActive = errors.New("post inactive")
	slugRegex       = regexp.MustCompile(`[^\w-]+`)
)

// PostService contains business logic for managing posts
type PostService struct {
	repo interfaces.IPostRepository
}

// NewPostService creates a new PostService with the given repository
func NewPostService(repo interfaces.IPostRepository) *PostService {
	return &PostService{repo: repo}
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

	post.Slug = slug
	createdPost, err := ps.repo.Create(ctx, post)
	if err != nil {
		log.Error("Failed to create post", slog.String("slug", post.Slug), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create post"), err)
	}

	log.Info("Post created", slog.String("id", createdPost.ID.String()), slog.String("slug", createdPost.Slug))
	return createdPost, nil
}

func (ps *PostService) ListPublishedAndPaginatedPosts(ctx context.Context, page, pageSize int) ([]model.PostPreview, int, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_published_posts_service")

	var posts []model.PostPreview
	var totalCount int
	var postsErr, countErr error

	done := make(chan bool, 2)

	go func() {
		posts, postsErr = ps.repo.ListPublished(ctx, page, pageSize)
		done <- true
	}()

	go func() {
		totalCount, countErr = ps.repo.CountPublished(ctx)
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

	baseSlug := generateSlug(t)
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

// generateSlug normalizes and sanitizes a string to create a URL-friendly slug
func generateSlug(t string) string {
	slug := removeAccents(t)
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRegex.ReplaceAllString(slug, "")

	return slug
}

// removeAccents removes diacritical marks (accents) from a string
func removeAccents(s string) string {
	/*
		1. Normalize the string to NFC (normalized form decomposed)
		2. Breaks accented letters into two runes: One for the letter and one for the accent
		 	 - Example: "São João"
			 - []rune{'S', 'a', '̃', 'o', ' ', 'J', 'o', '̃', 'a', 'o'}
	*/
	t := norm.NFD.String(s)

	result := make([]rune, 0, len(t))
	for _, r := range t {
		/*
			Mn -> Represents the unicode category "Mark, Nonspacing"
			- Thats include accents, cedillas, umlauts, tildes and any character
				that does not occupy it's own space - that is, combinable accents

			Is() checks if that rune belongs to the given category.
			If it's an accent (Mn), we ignore it with continue.
			If it's a letter or number, we add it to the rune slice.
		*/
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}
