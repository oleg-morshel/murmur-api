package posts_service

import (
	"context"
	"fmt"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

func (s *PostService) Delete(ctx context.Context, postID, userID int64) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("posts.Delete: %w", err)
	}

	if post.AuthorID != userID {
		return fmt.Errorf("posts.Delete: %w", core_errors.ErrForbidden)
	}

	return s.postRepo.Delete(ctx, postID)
}
