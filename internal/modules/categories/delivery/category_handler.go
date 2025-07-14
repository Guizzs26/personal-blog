package delivery

import (
	"log/slog"
	"net/http"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/contracts/dto"
	"github.com/Guizzs26/personal-blog/internal/modules/categories/service"
	"github.com/Guizzs26/personal-blog/pkg/httpx"
	"github.com/Guizzs26/personal-blog/pkg/validatorx"
	"github.com/go-playground/validator/v10"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (ch *CategoryHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromContext(ctx).WithGroup("create_category")

	req, err := httpx.Bind[dto.CreateCategoryRequest](r)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			httpx.WriteValidationErrors(w, validatorx.FormatValidationErrors(ve))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "Invalid request body")
		return
	}

	category := req.ToModel()
	createdCategory, err := ch.service.CreateCategory(ctx, category)
	if err != nil {
		log.Error("Failed to create category", slog.String("name", req.Name), slog.Any("error", err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "Failed to create post")
		return
	}

	log.Info("Category created", slog.String("id", createdCategory.ID.String()), slog.String("slug", createdCategory.Slug))

	res := dto.ToCategoryFullResponse(createdCategory)
	httpx.WriteJSON(w, http.StatusCreated, res)
}
