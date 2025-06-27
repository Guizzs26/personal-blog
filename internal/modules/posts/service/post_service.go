package service

import (
	"context"
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
	"github.com/mdobak/go-xerrors"
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

			log.Debug("Setting published_at for published post", slog.Time("published_at", now))
		}
	} else {
		post.PublishedAt = nil // just a guarantee
	}

	slug, err := ps.generateUniqueSlug(ctx, post.Title)
	if err != nil {
		log.Error("Failed to generate unique slug",
			slog.String("title", post.Title),
			slog.Any("error", err))

		return nil, xerrors.WithWrapper(xerrors.New("service: generate unique slug"), err)
	}

	log.Debug("Generated unique slug", slog.String("slug", slug))

	post.Slug = slug
	createdPost, err := ps.repo.Create(ctx, post)
	if err != nil {
		log.Error("Failed to create post in repository",
			slog.String("slug", post.Slug),
			slog.Any("repo_error", err))

		return nil, xerrors.WithWrapper(xerrors.New("service: create post"), err)
	}

	log.Info("Post created successfully in service",
		slog.String("post_id", createdPost.ID.String()),
		slog.String("slug", createdPost.Slug))

	return createdPost, nil
}

// generateUniqueSlug ensures that the generated slug is unique in the database
func (ps *PostService) generateUniqueSlug(ctx context.Context, t string) (string, error) {
	log := logger.GetLoggerFromContext(ctx)

	baseSlug := generateSlug(t)
	slug := baseSlug

	log.Debug("Generated base slug", slog.String("base_slug", baseSlug))

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

	log.Debug("Slug already exists, generating variations",
		slog.String("base_slug", baseSlug))
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
			log.Debug("Found unique slug after attempts",
				slog.String("final_slug", slug),
				slog.Int("attempts", i))
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

	re := regexp.MustCompile(`[^\w-]+`)
	slug = re.ReplaceAllString(slug, "")

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

func (ps *PostService) ListPublishedAndPaginatedPosts(ctx context.Context, page, pageSize int) ([]model.PostPreview, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("list_published_service")

	posts, err := ps.repo.ListPublished(ctx, page, pageSize)
	if err != nil {
		log.Error("Failed to list published posts", slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("service: list published posts"), err)
	}

	return posts, nil
}
