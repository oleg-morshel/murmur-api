package polls_postgres

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

func (r *PollRepository) Vote(ctx context.Context, vote *domain.PollVote) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PollRepository.Vote: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM poll_options WHERE id = $1 AND poll_id = $2)`,
		vote.OptionID, vote.PollID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("PollRepository.Vote: check option: %w", err)
	}
	if !exists {
		return fmt.Errorf("option %d does not belong to poll %d: %w", vote.OptionID, vote.PollID, core_errors.ErrBadRequest)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO poll_votes (poll_id, option_id, user_id) VALUES ($1, $2, $3)`,
		vote.PollID, vote.OptionID, vote.UserID,
	)
	if err != nil {
		return fmt.Errorf("PollRepository.Vote: insert vote: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE poll_options SET votes = votes + 1 WHERE id = $1`,
		vote.OptionID,
	)
	if err != nil {
		return fmt.Errorf("PollRepository.Vote: increment votes: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("PollRepository.Vote: commit: %w", err)
	}

	return nil
}
