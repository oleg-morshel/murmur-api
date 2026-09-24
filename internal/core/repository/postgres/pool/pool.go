package core_postgres_pool

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
	Begin(ctx context.Context) (pgx.Tx, error)
	OpTimeout() time.Duration
}

type ConnectionPool struct {
	*pgxpool.Pool
	timeout time.Duration
}

func NewConnectionPool(ctx context.Context, cfg Config, log *logger.Logger) (*ConnectionPool, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("postgres.NewConnectionPool: parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewConnectionPool: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres.NewConnectionPool: ping: %w", err)
	}

	log.Info("postgres connected", slog.String("host", cfg.Host), slog.String("db", cfg.Database))

	return &ConnectionPool{
		Pool:    pool,
		timeout: cfg.Timeout,
	}, nil
}

func (p *ConnectionPool) OpTimeout() time.Duration {
	return p.timeout
}
