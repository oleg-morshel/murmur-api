package polls_service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
	"github.com/oleg-morshel/murmur-api/internal/testutil"
)

func TestPollService_Vote(t *testing.T) {
	errDB := errors.New("db down")

	const (
		pollID   = int64(7)
		optionID = int64(3)
		userID   = int64(42)
	)

	tests := []struct {
		name             string
		hasVoted         bool
		hasVotedErr      error
		voteErr          error
		publishErr       error
		wantErr          error
		wantVoteCalls    int
		wantPublishCalls int
	}{
		{
			name:             "success",
			wantVoteCalls:    1,
			wantPublishCalls: 1,
		},
		{
			name:     "user already voted",
			hasVoted: true,
			wantErr:  core_errors.ErrConflict,
		},
		{
			name:        "has voted check fails",
			hasVotedErr: errDB,
			wantErr:     errDB,
		},
		{
			name:          "repository vote fails",
			voteErr:       errDB,
			wantErr:       errDB,
			wantVoteCalls: 1,
		},
		{
			name:             "publish fails but vote succeeds",
			publishErr:       errDB,
			wantVoteCalls:    1,
			wantPublishCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotVote          *domain.PollVote
				voteCalls        int
				publishCalls     int
				publishedSubject string
				publishedData    []byte
			)

			repo := &mockPollRepo{
				hasVotedFn: func(_ context.Context, gotPollID, gotUserID int64) (bool, error) {
					if gotPollID != pollID || gotUserID != userID {
						t.Errorf("HasVoted(%d, %d), want (%d, %d)", gotPollID, gotUserID, pollID, userID)
					}
					return tt.hasVoted, tt.hasVotedErr
				},
				voteFn: func(_ context.Context, v *domain.PollVote) error {
					voteCalls++
					gotVote = v
					return tt.voteErr
				},
			}
			publisher := &mockPublisher{
				publishFn: func(subject string, data []byte) error {
					publishCalls++
					publishedSubject = subject
					publishedData = data
					return tt.publishErr
				},
			}

			svc := NewPollService(repo, publisher)

			err := svc.Vote(testutil.NewContext(), pollID, optionID, userID)

			if voteCalls != tt.wantVoteCalls {
				t.Errorf("pollRepo.Vote calls = %d, want %d", voteCalls, tt.wantVoteCalls)
			}
			if publishCalls != tt.wantPublishCalls {
				t.Errorf("Publish calls = %d, want %d", publishCalls, tt.wantPublishCalls)
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotVote.PollID != pollID || gotVote.OptionID != optionID || gotVote.UserID != userID {
				t.Errorf("vote = %+v, want poll %d option %d user %d", gotVote, pollID, optionID, userID)
			}

			if publishedSubject != core_events.SubjectPollVoted {
				t.Errorf("subject = %q, want %q", publishedSubject, core_events.SubjectPollVoted)
			}

			var event core_events.PollVotedEvent
			if err := json.Unmarshal(publishedData, &event); err != nil {
				t.Fatalf("published data is not valid event JSON: %v", err)
			}
			if event.PollID != pollID || event.OptionID != optionID || event.UserID != userID {
				t.Errorf("event = %+v, want poll %d option %d user %d", event, pollID, optionID, userID)
			}
			if event.Timestamp.IsZero() {
				t.Error("event timestamp is zero")
			}
		})
	}
}
