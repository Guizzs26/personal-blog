package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

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
	slug, err := ps.generateUniqueSlug(post.Title)
	if err != nil {
		return nil, fmt.Errorf("service: failed to generate slug: %w", err)
	}

	post.Slug = slug
	createdPost, err := ps.repo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create post: %w", err)
	}

	return createdPost, nil
}

func (ps *PostService) generateUniqueSlug(t string) (string, error) {
	baseSlug := generateSlug(t)
	slug := baseSlug

	exists, err := ps.repo.ExistsBySlug(slug)
	if err != nil {
		return "", fmt.Errorf("failed to check slug existence: %w", err)
	}

	if !exists {
		return slug, nil
	}

	for i := 1; ; i++ {
		slug := fmt.Sprintf("%s-%d", baseSlug, i)

		exists, err := ps.repo.ExistsBySlug(slug)
		if err != nil {
			return "", fmt.Errorf("failed to check slug existence: %w", err)
		}

		if !exists {
			break
		}
	}
	return slug, nil
}

func generateSlug(t string) string {
	slug := removeAccents(t)
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")

	re := regexp.MustCompile(`[^\w-]+`)
	slug = re.ReplaceAllString(slug, "")

	return slug
}

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
