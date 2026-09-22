package posts_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
	FeedTTL      = 60 * time.Second
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

	log := logger.FromContext(ctx)
	cacheKey := fmt.Sprintf("feed:latest:%d:%d", limit, offset)

	data, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Warn("posts.List: cache get error", slog.Any("error", err))
	}
	if data != nil {
		var posts []*domain.Post
		if err := json.Unmarshal(data, &posts); err == nil {
			log.Debug("posts.List: cache hit", slog.String("key", cacheKey))
			return posts, nil
		}
	}

	log.Debug("posts.List: cache miss", slog.String("key", cacheKey))
	posts, err := s.postRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("posts.List: %w", err)
	}

	jsonData, err := json.Marshal(posts)
	if err != nil {
		log.Warn("posts.List: cache marshal error", slog.Any("error", err))
	} else {
		if err := s.cache.Set(ctx, cacheKey, jsonData, FeedTTL); err != nil {
			log.Warn("posts.List: cache set error", slog.Any("error", err))
		}
	}

	return posts, nil
}
