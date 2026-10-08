package polls_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type mockPollRepo struct {
	createPollFn  func(ctx context.Context, poll *domain.Poll) (int64, error)
	getByPostIDFn func(ctx context.Context, postID int64) (*domain.Poll, error)
	voteFn        func(ctx context.Context, vote *domain.PollVote) error
	hasVotedFn    func(ctx context.Context, pollID, userID int64) (bool, error)
}

func (m *mockPollRepo) CreatePoll(ctx context.Context, poll *domain.Poll) (int64, error) {
	return m.createPollFn(ctx, poll)
}

func (m *mockPollRepo) GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error) {
	return m.getByPostIDFn(ctx, postID)
}

func (m *mockPollRepo) Vote(ctx context.Context, vote *domain.PollVote) error {
	return m.voteFn(ctx, vote)
}

func (m *mockPollRepo) HasVoted(ctx context.Context, pollID, userID int64) (bool, error) {
	return m.hasVotedFn(ctx, pollID, userID)
}

type mockPublisher struct {
	publishFn func(subject string, data []byte) error
}

func (m *mockPublisher) Publish(subject string, data []byte) error {
	return m.publishFn(subject, data)
}
