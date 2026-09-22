package posts_service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

func (s *PostService) Create(ctx context.Context, authorID int64, content string, anonymous bool) (*domain.Post, error) {
	log := logger.FromContext(ctx)
	allowed, err := s.rateLimiter.Allow(ctx, authorID)
	if err != nil {
		log.Warn("posts.Create: rate limiter error", slog.Any("error", err))
	}

	if !allowed {
		return nil, core_errors.ErrRateLimited
	}

	post := &domain.Post{
		AuthorID:  authorID,
		Content:   content,
		Anonymous: anonymous,
	}

	postID, err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("posts.Create: %w", err)
	}

	if err := s.cache.DeleteByPattern(ctx, "feed:latest:*"); err != nil {
		log.Warn("posts.Create: cache invalidate error", slog.Any("error", err))
	}

	return s.postRepo.GetByID(ctx, postID)
}
