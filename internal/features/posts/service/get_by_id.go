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

const PostTTL = 5 * time.Minute

func (s *PostService) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	log := logger.FromContext(ctx)
	cacheKey := fmt.Sprintf("post:%d", id)

	data, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Warn("posts.GetByID: cache get error", slog.Any("error", err))
	}
	if data != nil {
		var post domain.Post
		if err := json.Unmarshal(data, &post); err == nil {
			log.Debug("posts.GetByID: cache hit", slog.String("key", cacheKey))
			return &post, nil
		}
	}

	log.Debug("posts.GetByID: cache miss", slog.String("key", cacheKey))
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("posts.GetByID: %w", err)
	}

	jsonData, err := json.Marshal(post)
	if err != nil {
		log.Warn("posts.GetByID: cache marshal error", slog.Any("error", err))
	} else {
		if err := s.cache.Set(ctx, cacheKey, jsonData, PostTTL); err != nil {
			log.Warn("posts.GetByID: cache set error", slog.Any("error", err))
		}
	}

	return post, nil
}
