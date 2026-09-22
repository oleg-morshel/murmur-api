package posts_service

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

func (s *PostService) Update(ctx context.Context, postID, userID int64, content string, anonymous bool) (*domain.Post, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("posts.Update: %w", err)
	}

	if post.AuthorID != userID {
		return nil, core_errors.ErrForbidden
	}

	post.Content = content
	post.Anonymous = anonymous

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("posts.Update: %w", err)
	}

	return post, nil
}
