package polls_service

import (
	"context"
	"errors"
	"testing"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func TestPollService_CreatePoll(t *testing.T) {
	errDB := errors.New("db down")

	const postID = int64(1)
	const question = "Tabs or spaces?"
	options := []string{"tabs", "spaces", "both"}

	created := &domain.Poll{ID: 5, PostID: postID, Question: question}

	tests := []struct {
		name      string
		createErr error
		getErr    error
		wantErr   error
	}{
		{name: "success"},
		{name: "create fails", createErr: errDB, wantErr: errDB},
		{name: "reading created poll fails", getErr: errDB, wantErr: errDB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPoll *domain.Poll

			repo := &mockPollRepo{
				createPollFn: func(_ context.Context, p *domain.Poll) (int64, error) {
					gotPoll = p
					if tt.createErr != nil {
						return 0, tt.createErr
					}
					return created.ID, nil
				},
				getByPostIDFn: func(_ context.Context, id int64) (*domain.Poll, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return created, nil
				},
			}

			svc := NewPollService(repo, &mockPublisher{})

			poll, err := svc.CreatePoll(context.Background(), postID, question, options)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if poll != nil {
					t.Errorf("poll = %+v, want nil on error", poll)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if poll != created {
				t.Errorf("poll = %+v, want %+v", poll, created)
			}

			if gotPoll.PostID != postID || gotPoll.Question != question {
				t.Errorf("poll passed to repo = %+v", gotPoll)
			}
			if len(gotPoll.Options) != len(options) {
				t.Fatalf("options = %d, want %d", len(gotPoll.Options), len(options))
			}
			for i, opt := range gotPoll.Options {
				if opt.Text != options[i] {
					t.Errorf("option[%d] = %q, want %q", i, opt.Text, options[i])
				}
			}
		})
	}
}
