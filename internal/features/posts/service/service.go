package posts_service

import (
	"context"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id int64) error
}

type PostCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
}

type PollProvider interface {
	GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, userID int64) (bool, error)
}

type EventPublisher interface {
	Publish(subject string, data []byte) error
}

type PostService struct {
	postRepo       PostRepository
	cache          PostCache
	rateLimiter    RateLimiter
	pollProvider   PollProvider
	eventPublisher EventPublisher
}

func NewPostService(postRepo PostRepository, cache PostCache, rateLimiter RateLimiter, pollProvider PollProvider, eventPublisher EventPublisher) *PostService {
	return &PostService{
		postRepo:       postRepo,
		cache:          cache,
		rateLimiter:    rateLimiter,
		pollProvider:   pollProvider,
		eventPublisher: eventPublisher,
	}
}
