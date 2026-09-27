package polls_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
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

	log := logger.FromContext(ctx)
	event := core_events.PollVotedEvent{
		Type:      core_events.SubjectPollVoted,
		PollID:    pollID,
		OptionID:  optionID,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(event); err != nil {
		log.Warn("polls.Vote: event marshal error", slog.Any("error", err))
	} else if err := s.eventPublisher.Publish(core_events.SubjectPollVoted, data); err != nil {
		log.Warn("polls.Vote: event publish error", slog.Any("error", err))
	}

	return nil
}
