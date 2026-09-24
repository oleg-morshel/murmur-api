package polls_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

func (r *PollRepository) GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error) {
	poll := &domain.Poll{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, post_id, question, created_at, updated_at FROM polls WHERE post_id = $1`,
		postID,
	).Scan(&poll.ID, &poll.PostID, &poll.Question, &poll.CreatedAt, &poll.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("poll not found for post %d: %w", postID, core_errors.ErrNotFound)
		}
		return nil, fmt.Errorf("PollRepository.GetByPostID: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, poll_id, text, votes, position FROM poll_options WHERE poll_id = $1 ORDER BY position`,
		poll.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("PollRepository.GetByPostID: options: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		opt := &domain.PollOption{}
		if err := rows.Scan(&opt.ID, &opt.PollID, &opt.Text, &opt.Votes, &opt.Position); err != nil {
			return nil, fmt.Errorf("PollRepository.GetByPostID: scan option: %w", err)
		}
		poll.Options = append(poll.Options, opt)
	}

	return poll, nil
}
