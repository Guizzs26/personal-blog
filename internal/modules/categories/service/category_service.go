package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/contracts/interfaces"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
	"github.com/Guizzs26/personal-blog/pkg/slug"
	"github.com/mdobak/go-xerrors"
)

type CategoryService struct {
	repo interfaces.ICategoryRepository
}

func NewCategoryService(repo interfaces.ICategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (cs *CategoryService) CreateCategory(ctx context.Context, category model.Category) (*model.Category, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_category_service")

	slug, err := cs.generateUniqueSlug(ctx, category.Name)
	if err != nil {
		log.Error("Failed to generate unique slug", slog.String("name", category.Name), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to generate unique slug"), err)
	}

	category.Slug = slug
	createdCategory, err := cs.repo.Create(ctx, category)
	if err != nil {
		log.Error("Failed to create category", slog.String("slug", category.Slug), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create category"), err)
	}

	log.Info("Category created", slog.String("id", createdCategory.ID.String()), slog.String("slug", createdCategory.Slug))
	return createdCategory, nil
}

func (cs *CategoryService) generateUniqueSlug(ctx context.Context, n string) (string, error) {
	log := logger.GetLoggerFromContext(ctx)

	baseSlug := slug.GenerateSlug(n)
	slug := baseSlug

	exists, err := cs.repo.ExistsBySlug(ctx, slug)
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

		exists, err := cs.repo.ExistsBySlug(ctx, slug)
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
