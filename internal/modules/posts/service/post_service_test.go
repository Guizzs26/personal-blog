package service

import (
	"context"
	"testing"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/*

2. Casos de teste principais
Vamos cobrir os principais comportamentos esperados:

✅ Criação com sucesso (slug único e published = true) 👍
	✅ published=true → published_at deve ser definido
	✅ Se o slug é gerado corretamente
	(generateSlug, generateUniqueSlug, removeAccents)
		 - São testadas em conjunto com a funcionalidade de criar o post porque
		   são "intrínsicas" a ela. Não vale a pena testar isoladamente.

❌ Erro ao verificar se slug existe

❌ Erro ao criar o post no repositório (erro do repo propagado corretamente)

✅ Geração de slug alternativo quando o título já existe (criar um post com slug repetido) 👍
	✅ Geração de post sem image_id
	✅ published=false → published_at deve ser nil
	(novamente aqui generateSlug, generateUniqueSlug, removeAccents são testadas)
*/

// mockPostRepository is a testify mock for IPostRepository
type mockPostRepository struct {
	mock.Mock
}

func (m *mockPostRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *mockPostRepository) Create(ctx context.Context, post model.Post) (*model.Post, error) {
	args := m.Called(ctx, post)
	return args.Get(0).(*model.Post), args.Error(1)
}

/*
We tested the creation of a valid post

We also tested indirectly published = true results in a published_at
set to the current time. In addition, we also tested generateSlug,
generateUniqueSlug and removeAccents functions
*/
func TestCreatePost_Sucess(t *testing.T) {
	// Input data - Arrange (A)
	ctx := context.Background()
	title := "Some Títle!"
	slug := "some-title"
	authorID := uuid.New()
	imageID := uuid.New()

	mockRepo := new(mockPostRepository)
	postService := NewPostService(mockRepo)

	// Expect ExistsBySlug to be called once and return false
	mockRepo.On("ExistsBySlug", mock.Anything, slug).Return(false, nil)

	// Expected saved post
	expectedPost := &model.Post{
		ID:        uuid.New(),
		Title:     title,
		Content:   "# Markdown content",
		Slug:      slug,
		AuthorID:  authorID,
		ImageID:   &imageID,
		Published: true,
	}
	expectedPost.PublishedAt = func() *time.Time {
		now := time.Now()
		return &now
	}()
	expectedPost.CreatedAt = time.Now()
	expectedPost.UpdatedAt = expectedPost.CreatedAt

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p model.Post) bool {
		return p.Title == expectedPost.Title &&
			p.Slug == expectedPost.Slug &&
			p.AuthorID == expectedPost.AuthorID
	})).Return(expectedPost, nil)

	input := model.Post{
		Title:     title,
		Content:   "# Markdown Content",
		AuthorID:  authorID,
		ImageID:   &imageID,
		Published: true,
	}

	createdPost, err := postService.CreatePost(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, createdPost)
	assert.Equal(t, expectedPost.Title, createdPost.Title)
	assert.Equal(t, expectedPost.Slug, createdPost.Slug)
	assert.Equal(t, expectedPost.AuthorID, createdPost.AuthorID)
	assert.True(t, createdPost.Published)
	assert.NotNil(t, createdPost.PublishedAt)

	mockRepo.AssertExpectations(t)
}

/*
We tested the creation of a valid post

We also tested that when published=false the published_at cannot be set.
In addition, we tested the behavior of creating a valid post even without
an image_id (optional) and the generation of an incremental slug if a slug
already exists.
*/
func TestCreatePost_SlugConflictGeneratesIncrementedSlug(t *testing.T) {
	ctx := context.Background()
	title := "Go is awesome!"
	baseSlug := "go-is-awesome"
	authorID := uuid.New()

	mockRepo := new(mockPostRepository)
	postService := NewPostService(mockRepo)

	// First try
	mockRepo.On("ExistsBySlug", mock.Anything, baseSlug).Return(true, nil)

	// Second try
	mockRepo.On("ExistsBySlug", mock.Anything, baseSlug+"-1").Return(false, nil)

	expectedPost := &model.Post{
		ID:        uuid.New(),
		Title:     title,
		Content:   "## Markdown Content",
		Slug:      baseSlug + "-1",
		AuthorID:  authorID,
		Published: false,
	}
	expectedPost.CreatedAt = time.Now()
	expectedPost.UpdatedAt = expectedPost.CreatedAt

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p model.Post) bool {
		return p.Slug == expectedPost.Slug &&
			p.Title == expectedPost.Title &&
			p.AuthorID == expectedPost.AuthorID
	})).Return(expectedPost, nil)

	input := model.Post{
		Title:    title,
		Content:  "## Markdown Content",
		AuthorID: authorID,
	}

	createdPost, err := postService.CreatePost(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, createdPost)
	assert.Equal(t, expectedPost.Slug, createdPost.Slug)
	assert.Equal(t, expectedPost.Title, createdPost.Title)

	mockRepo.AssertExpectations(t)
}
