package posts_service

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

func TestPostService_Create(t *testing.T) {
	errDB := errors.New("db down")

	const (
		authorID  = int64(5)
		postID    = int64(77)
		content   = "hello, murmur"
		anonymous = true
	)

	tests := []struct {
		name             string
		allowed          bool
		limiterErr       error
		createErr        error
		cacheErr         error
		publishErr       error
		wantErr          error
		wantCreateCalls  int
		wantCacheCalls   int
		wantPublishCalls int
	}{
		{
			name:             "success",
			allowed:          true,
			wantCreateCalls:  1,
			wantCacheCalls:   1,
			wantPublishCalls: 1,
		},
		{
			name:    "rate limited",
			allowed: false,
			wantErr: core_errors.ErrRateLimited,
		},
		{
			name:             "limiter error is only logged, allowed flag decides",
			allowed:          true,
			limiterErr:       errDB,
			wantCreateCalls:  1,
			wantCacheCalls:   1,
			wantPublishCalls: 1,
		},
		{
			name:            "repository create fails",
			allowed:         true,
			createErr:       errDB,
			wantErr:         errDB,
			wantCreateCalls: 1,
		},
		{
			name:             "cache and publish errors do not fail creation",
			allowed:          true,
			cacheErr:         errDB,
			publishErr:       errDB,
			wantCreateCalls:  1,
			wantCacheCalls:   1,
			wantPublishCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := &domain.Post{AuthorID: authorID, Content: content, Anonymous: anonymous}

			var (
				gotPost          *domain.Post
				limiterUser      int64
				createCalls      int
				cacheCalls       int
				publishCalls     int
				invalidatedGlob  string
				publishedSubject string
				publishedData    []byte
			)

			repo := &mockPostRepo{
				createFn: func(_ context.Context, p *domain.Post) (int64, error) {
					createCalls++
					gotPost = p
					if tt.createErr != nil {
						return 0, tt.createErr
					}
					return postID, nil
				},
				getByIDFn: func(_ context.Context, id int64) (*domain.Post, error) {
					if id != postID {
						t.Errorf("GetByID(%d), want %d", id, postID)
					}
					return stored, nil
				},
			}
			cache := &mockPostCache{
				deleteByPatternFn: func(_ context.Context, pattern string) error {
					cacheCalls++
					invalidatedGlob = pattern
					return tt.cacheErr
				},
			}
			limiter := &mockRateLimiter{
				allowFn: func(_ context.Context, userID int64) (bool, error) {
					limiterUser = userID
					return tt.allowed, tt.limiterErr
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

			svc := NewPostService(repo, cache, limiter, &mockPollProvider{}, publisher)

			post, err := svc.Create(testutil.NewContext(), authorID, content, anonymous)

			if limiterUser != authorID {
				t.Errorf("rate limiter checked user %d, want %d", limiterUser, authorID)
			}
			if createCalls != tt.wantCreateCalls {
				t.Errorf("postRepo.Create calls = %d, want %d", createCalls, tt.wantCreateCalls)
			}
			if cacheCalls != tt.wantCacheCalls {
				t.Errorf("cache.DeleteByPattern calls = %d, want %d", cacheCalls, tt.wantCacheCalls)
			}
			if publishCalls != tt.wantPublishCalls {
				t.Errorf("Publish calls = %d, want %d", publishCalls, tt.wantPublishCalls)
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if post != nil {
					t.Errorf("post = %+v, want nil on error", post)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if post != stored {
				t.Errorf("post = %+v, want %+v", post, stored)
			}

			if gotPost.AuthorID != authorID || gotPost.Content != content || gotPost.Anonymous != anonymous {
				t.Errorf("post passed to repo = %+v", gotPost)
			}

			if invalidatedGlob != "feed:latest:*" {
				t.Errorf("invalidated pattern = %q, want %q", invalidatedGlob, "feed:latest:*")
			}

			if publishedSubject != core_events.SubjectPostCreated {
				t.Errorf("subject = %q, want %q", publishedSubject, core_events.SubjectPostCreated)
			}
			var event core_events.PostCreatedEvent
			if err := json.Unmarshal(publishedData, &event); err != nil {
				t.Fatalf("published data is not valid event JSON: %v", err)
			}
			if event.PostID != postID || event.AuthorID != authorID {
				t.Errorf("event = %+v, want post %d author %d", event, postID, authorID)
			}
		})
	}
}
