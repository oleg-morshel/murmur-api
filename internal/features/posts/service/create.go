package posts_service

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func (s *PostService) Create(ctx context.Context, authorID int64, content string, anonymous bool) (*domain.Post, error) {
	post := &domain.Post{
		AuthorID:  authorID,
		Content:   content,
		Anonymous: anonymous,
	}

	id, err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("posts.Create: %w", err)
	}

	created, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("posts.Create: fetch created: %w", err)
	}

	return created, nil
}
