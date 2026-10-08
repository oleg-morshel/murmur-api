package posts_service

import (
	"context"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type mockPostRepo struct {
	createFn  func(ctx context.Context, post *domain.Post) (int64, error)
	getByIDFn func(ctx context.Context, id int64) (*domain.Post, error)
	listFn    func(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	updateFn  func(ctx context.Context, post *domain.Post) error
	deleteFn  func(ctx context.Context, id int64) error
}

func (m *mockPostRepo) Create(ctx context.Context, post *domain.Post) (int64, error) {
	return m.createFn(ctx, post)
}

func (m *mockPostRepo) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockPostRepo) List(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	return m.listFn(ctx, limit, offset)
}

func (m *mockPostRepo) Update(ctx context.Context, post *domain.Post) error {
	return m.updateFn(ctx, post)
}

func (m *mockPostRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

type mockPostCache struct {
	getFn             func(ctx context.Context, key string) ([]byte, error)
	setFn             func(ctx context.Context, key string, value []byte, ttl time.Duration) error
	deleteFn          func(ctx context.Context, keys ...string) error
	deleteByPatternFn func(ctx context.Context, pattern string) error
}

func (m *mockPostCache) Get(ctx context.Context, key string) ([]byte, error) {
	return m.getFn(ctx, key)
}

func (m *mockPostCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return m.setFn(ctx, key, value, ttl)
}

func (m *mockPostCache) Delete(ctx context.Context, keys ...string) error {
	return m.deleteFn(ctx, keys...)
}

func (m *mockPostCache) DeleteByPattern(ctx context.Context, pattern string) error {
	return m.deleteByPatternFn(ctx, pattern)
}

type mockRateLimiter struct {
	allowFn func(ctx context.Context, userID int64) (bool, error)
}

func (m *mockRateLimiter) Allow(ctx context.Context, userID int64) (bool, error) {
	return m.allowFn(ctx, userID)
}

type mockPollProvider struct {
	getByPostIDFn func(ctx context.Context, postID int64) (*domain.Poll, error)
}

func (m *mockPollProvider) GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error) {
	return m.getByPostIDFn(ctx, postID)
}

type mockPublisher struct {
	publishFn func(subject string, data []byte) error
}

func (m *mockPublisher) Publish(subject string, data []byte) error {
	return m.publishFn(subject, data)
}
