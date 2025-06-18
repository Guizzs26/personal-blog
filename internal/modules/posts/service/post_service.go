package service

import (
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/posts/model"
	"github.com/Guizzs26/personal-blog/internal/modules/posts/repository"
)

type PostService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (ps *PostService) CreatePost(post model.Post) (*model.Post, error) {
	createdPost, err := ps.repo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create post: %w", err)
	}

	return createdPost, nil
}
