package posts_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("RedisCache.Get: %w", err)
	}
	return data, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("RedisCache.Set: %w", err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("RedisCache.Delete: %w", err)
	}
	return nil
}

func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("RedisCache.DeleteByPattern: scan: %w", err)
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("RedisCache.DeleteByPattern: del: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
