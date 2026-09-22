package posts_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("ratelimit:posts:create:%d", userID)

	count, err := rl.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("RateLimiter.Allow: incr: %w", err)
	}

	if count == 1 {
		rl.rdb.Expire(ctx, key, rl.window)
	}

	return count <= int64(rl.limit), nil
}
