package polls_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type PollRepository interface {
	CreatePoll(ctx context.Context, poll *domain.Poll) (int64, error)
	GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error)
	Vote(ctx context.Context, vote *domain.PollVote) error
	HasVoted(ctx context.Context, pollID, userID int64) (bool, error)
}

type EventPublisher interface {
	Publish(subject string, data []byte) error
}

type PollService struct {
	pollRepo       PollRepository
	eventPublisher EventPublisher
}

func NewPollService(pollRepo PollRepository, eventPublisher EventPublisher) *PollService {
	return &PollService{
		pollRepo:       pollRepo,
		eventPublisher: eventPublisher,
	}
}
