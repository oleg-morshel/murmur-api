package auth_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
)

type TokenRepository struct {
	pool core_postgres_pool.Pool
}

func NewTokenRepository(pool core_postgres_pool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

func (r *TokenRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(localCtx, query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt)

	if err != nil {
		return fmt.Errorf("TokenRepository.Save: %w", err)
	}

	return nil
}

func (r *TokenRepository) DeleteByHash(ctx context.Context, tokenHash string) (bool, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	result, err := r.pool.Exec(localCtx, query, tokenHash)

	if err != nil {
		return false, fmt.Errorf("TokenRepository.DeleteByHash: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM refresh_tokens WHERE expires_at < now()`
	tag, err := r.pool.Exec(localCtx, query)
	if err != nil {
		return 0, fmt.Errorf("TokenRepository.DeleteExpired: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *TokenRepository) ConsumeByHash(ctx context.Context, tokenHash string) (int64, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        DELETE FROM refresh_tokens
        WHERE token_hash = $1 AND expires_at > now()
        RETURNING user_id`

	var userID int64
	err := r.pool.QueryRow(localCtx, query, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, core_errors.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("TokenRepository.ConsumeByHash: %w", err)
	}
	return userID, nil
}
