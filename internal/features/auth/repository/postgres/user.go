package auth_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
)

type AuthRepository struct {
	pool core_postgres_pool.Pool
}

func NewAuthRepository(pool core_postgres_pool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) Create(ctx context.Context, user *domain.User) (int64, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id`

	var id int64
	err := r.pool.QueryRow(localCtx, query, user.Username, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, fmt.Errorf("%w: user already exists", core_errors.ErrConflict)
		}
		return 0, fmt.Errorf("UserRepository.Create: %w", err)
	}
	return id, nil
}

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(localCtx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", core_errors.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}

	return user, nil
}

func (r *AuthRepository) GetById(ctx context.Context, id int64) (*domain.User, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(localCtx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("UserRepository.GetById: %w", core_errors.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetById: %w", err)
	}

	return user, nil
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
