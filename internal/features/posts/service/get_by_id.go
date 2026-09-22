package posts_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func (s *PostService) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	return s.postRepo.GetByID(ctx, id)
}
