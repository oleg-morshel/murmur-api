package core_redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/oleg-morshel/murmur-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(ctx context.Context, cfg Config, log *logger.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis.NewClient: ping: %w", err)
	}

	log.Info("redis connected", slog.String("addr", cfg.Addr()), slog.Int("db", cfg.DB))

	return &Client{rdb: rdb}, nil
}

func (c *Client) RDB() *redis.Client {
	return c.rdb
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
