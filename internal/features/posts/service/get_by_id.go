package posts_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

const PostTTL = 5 * time.Minute

func (s *PostService) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	log := logger.FromContext(ctx)
	cacheKey := fmt.Sprintf("post:%d", id)

	var post *domain.Post

	data, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Warn("posts.GetByID: cache get error", slog.Any("error", err))
	}

	if data != nil {
		var cached domain.Post
		if err := json.Unmarshal(data, &cached); err == nil {
			log.Debug("posts.GetByID: cache hit", slog.String("key", cacheKey))
			post = &cached
		}
	}

	if post == nil {
		log.Debug("posts.GetByID: cache miss", slog.String("key", cacheKey))
		post, err = s.postRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("posts.GetByID: %w", err)
		}

		jsonData, err := json.Marshal(post)
		if err != nil {
			log.Warn("posts.GetByID: cache marshal error", slog.Any("error", err))
		} else if err := s.cache.Set(ctx, cacheKey, jsonData, PostTTL); err != nil {
			log.Warn("posts.GetByID: cache set error", slog.Any("error", err))
		}
	}

	poll, err := s.pollProvider.GetByPostID(ctx, post.ID)
	if err != nil && !errors.Is(err, core_errors.ErrNotFound) {
		log.Warn("posts.GetByID: poll fetch error", slog.Any("error", err))
	} else if poll != nil {
		post.Poll = poll
	}

	return post, nil
}
