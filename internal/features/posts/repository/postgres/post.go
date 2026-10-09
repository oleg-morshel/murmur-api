package posts_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
)

type PostRepository struct {
	pool core_postgres_pool.Pool
}

func NewPostRepository(pool core_postgres_pool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) (int64, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO posts (author_id, content, anonymous)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(localCtx, query, post.AuthorID, post.Content, post.Anonymous).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("PostRepository.Create: %w", err)
	}

	return post.ID, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT p.id, p.author_id, p.content, p.anonymous, p.created_at, p.updated_at,
		       u.id, u.username
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = $1`

	post := &domain.Post{Author: &domain.User{}}
	err := r.pool.QueryRow(localCtx, query, id).Scan(
		&post.ID, &post.AuthorID, &post.Content, &post.Anonymous, &post.CreatedAt, &post.UpdatedAt,
		&post.Author.ID, &post.Author.Username,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("PostRepository.GetByID: %w", core_errors.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("PostRepository.GetByID: %w", err)
	}

	return post, nil
}

func (r *PostRepository) List(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT p.id, p.author_id, p.content, p.anonymous, p.created_at, p.updated_at,
		       u.id, u.username
		FROM posts p
		JOIN users u ON u.id = p.author_id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(localCtx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("PostRepository.List: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		post := &domain.Post{Author: &domain.User{}}
		if err := rows.Scan(
			&post.ID, &post.AuthorID, &post.Content, &post.Anonymous, &post.CreatedAt, &post.UpdatedAt,
			&post.Author.ID, &post.Author.Username,
		); err != nil {
			return nil, fmt.Errorf("PostRepository.List: scan: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PostRepository.List: rows: %w", err)
	}

	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE posts
		SET content = $1, anonymous = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at`

	err := r.pool.QueryRow(localCtx, query, post.Content, post.Anonymous, post.ID).
		Scan(&post.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("PostRepository.Update: %w", core_errors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("PostRepository.Update: %w", err)
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id int64) error {
	localCtx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM posts WHERE id = $1`

	tag, err := r.pool.Exec(localCtx, query, id)
	if err != nil {
		return fmt.Errorf("PostRepository.Delete: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("PostRepository.Delete: %w", core_errors.ErrNotFound)
	}

	return nil
}
