package polls_postgres

import (
	"context"
	"fmt"
)

func (r *PollRepository) HasVoted(ctx context.Context, pollID, userID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM poll_votes WHERE poll_id = $1 AND user_id = $2)`,
		pollID, userID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("PollRepository.HasVoted: %w", err)
	}

	return exists, nil
}
