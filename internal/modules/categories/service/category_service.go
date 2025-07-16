package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/contracts/interfaces"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/model"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/repository"
	"github.com/Guizzs26/personal-blog/pkg/slug"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryIsActive = errors.New("category inactive")
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

func (cs *CategoryService) ListActiveAndPaginatedCategories(ctx context.Context, page, pageSize int) (*[]model.Category, int, error) {
	categories, err := cs.repo.ListActives(ctx, page, pageSize)
	if err != nil {
		return nil, 0, xerrors.WithWrapper(xerrors.New("failed to list active categories"), err)
	}

	totalCount, err := cs.repo.CountActives(ctx)
	if err != nil {
		return nil, 0, xerrors.WithWrapper(xerrors.New("failed to count active categories"), err)
	}

	return categories, totalCount, nil
}

func (cs *CategoryService) UpdateCategoryByID(ctx context.Context, id uuid.UUID, name string) (*model.Category, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("category_service")

	slug, err := cs.generateUniqueSlug(ctx, name)
	if err != nil {
		return nil, xerrors.WithWrapper(xerrors.New("failed to generate slug"), err)
	}

	updatedCategory, err := cs.repo.UpdateByID(ctx, id, name, slug)
	if errors.Is(err, repository.ErrResourceNotFound) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		log.Error("Failed to update category", slog.String("id", id.String()), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to update category"), err)
	}

	log.Info("Post updated", slog.String("id", updatedCategory.ID.String()))
	return updatedCategory, nil
}

func (cs *CategoryService) SetCategoryActive(ctx context.Context, id uuid.UUID, active bool) (*model.Category, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("set_active_category_service")

	category, err := cs.repo.SetActive(ctx, id, active)
	if err != nil {
		log.Error("Failed to update category status", slog.String("id", id.String()), slog.Bool("active", active), slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to update category status"), err)
	}

	log.Info("Category status updated",
		slog.String("id", category.ID.String()),
		slog.String("slug", category.Slug),
		slog.Bool("active", active))
	return category, nil
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
