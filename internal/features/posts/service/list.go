package posts_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

func (s *PostService) List(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.postRepo.List(ctx, limit, offset)
}
