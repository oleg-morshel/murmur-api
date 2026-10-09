package posts_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

func (s *PostService) Delete(ctx context.Context, postID, userID int64) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("posts.Delete: %w", err)
	}

	if post.AuthorID != userID {
		return core_errors.ErrForbidden
	}

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("posts.Delete: %w", err)
	}

	log := logger.FromContext(ctx)
	if err := s.cache.Delete(ctx, fmt.Sprintf("post:%d", postID)); err != nil {
		log.Warn("posts.Delete: cache invalidate error", slog.Any("error", err))
	}
	if err := s.cache.DeleteByPattern(ctx, "feed:latest:*"); err != nil {
		log.Warn("posts.Delete: cache invalidate error", slog.Any("error", err))
	}

	event := core_events.PostDeletedEvent{
		Type:      core_events.SubjectPostDeleted,
		PostID:    postID,
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(event); err != nil {
		log.Warn("posts.Delete: event marshal error", slog.Any("error", err))
	} else if err := s.eventPublisher.Publish(core_events.SubjectPostDeleted, data); err != nil {
		log.Warn("posts.Delete: event publish error", slog.Any("error", err))
	}

	return nil
}
