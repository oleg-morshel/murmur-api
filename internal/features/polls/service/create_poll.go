package polls_service

import (
	"context"
	"fmt"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func (s *PollService) CreatePoll(ctx context.Context, postID int64, question string, options []string) (*domain.Poll, error) {
	poll := &domain.Poll{
		PostID:   postID,
		Question: question,
	}

	for _, text := range options {
		poll.Options = append(poll.Options, &domain.PollOption{Text: text})
	}

	_, err := s.pollRepo.CreatePoll(ctx, poll)
	if err != nil {
		return nil, fmt.Errorf("polls.CreatePoll: %w", err)
	}

	return s.pollRepo.GetByPostID(ctx, postID)
}
