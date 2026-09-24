package polls_service

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
)

func (s *PollService) Vote(ctx context.Context, pollID, optionID, userID int64) error {
	voted, err := s.pollRepo.HasVoted(ctx, pollID, userID)
	if err != nil {
		return fmt.Errorf("polls.Vote: %w", err)
	}
	if voted {
		return fmt.Errorf("user already voted in this poll: %w", core_errors.ErrConflict)
	}

	vote := &domain.PollVote{
		PollID:   pollID,
		OptionID: optionID,
		UserID:   userID,
	}

	if err := s.pollRepo.Vote(ctx, vote); err != nil {
		return fmt.Errorf("polls.Vote: %w", err)
	}

	return nil
}
