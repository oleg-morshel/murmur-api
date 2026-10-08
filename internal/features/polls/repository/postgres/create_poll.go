package polls_postgres

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func (r *PollRepository) CreatePoll(ctx context.Context, poll *domain.Poll) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("PollRepository.CreatePoll: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pollID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO polls (post_id, question) VALUES ($1, $2) RETURNING id`,
		poll.PostID, poll.Question,
	).Scan(&pollID)
	if err != nil {
		return 0, fmt.Errorf("PollRepository.CreatePoll: insert poll: %w", err)
	}

	for i, opt := range poll.Options {
		_, err = tx.Exec(ctx,
			`INSERT INTO poll_options (poll_id, text, position) VALUES ($1, $2, $3)`,
			pollID, opt.Text, i,
		)
		if err != nil {
			return 0, fmt.Errorf("PollRepository.CreatePoll: insert option: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("PollRepository.CreatePoll: commit: %w", err)
	}

	return pollID, nil
}
