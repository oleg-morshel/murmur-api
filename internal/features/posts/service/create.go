package posts_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
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

	event := core_events.PostCreatedEvent{
		Type:      core_events.SubjectPostCreated,
		PostID:    postID,
		AuthorID:  authorID,
		Anonymous: anonymous,
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(event); err != nil {
		log.Warn("posts.Create: event marshal error", slog.Any("error", err))
	} else if err := s.eventPublisher.Publish(core_events.SubjectPostCreated, data); err != nil {
		log.Warn("posts.Create: event publish error", slog.Any("error", err))
	}

	return s.postRepo.GetByID(ctx, postID)
}
