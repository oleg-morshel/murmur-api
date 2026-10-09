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

func TestPostService_Delete(t *testing.T) {
	errDB := errors.New("db down")

	const (
		postID = int64(10)
		owner  = int64(1)
		other  = int64(2)
	)

	tests := []struct {
		name             string
		postAuthorID     int64
		userID           int64
		getErr           error
		deleteErr        error
		cacheErr         error
		publishErr       error
		wantErr          error
		wantDeleteCalls  int
		wantCacheCalls   int
		wantPublishCalls int
	}{
		{
			name:             "owner deletes own post",
			postAuthorID:     owner,
			userID:           owner,
			wantDeleteCalls:  1,
			wantCacheCalls:   1,
			wantPublishCalls: 1,
		},
		{
			name:    "post not found",
			userID:  owner,
			getErr:  core_errors.ErrNotFound,
			wantErr: core_errors.ErrNotFound,
		},
		{
			name:         "someone else's post is forbidden",
			postAuthorID: owner,
			userID:       other,
			wantErr:      core_errors.ErrForbidden,
		},
		{
			name:            "repository delete fails",
			postAuthorID:    owner,
			userID:          owner,
			deleteErr:       errDB,
			wantErr:         errDB,
			wantDeleteCalls: 1,
		},
		{
			name:             "cache and publish errors do not fail deletion",
			postAuthorID:     owner,
			userID:           owner,
			cacheErr:         errDB,
			publishErr:       errDB,
			wantDeleteCalls:  1,
			wantCacheCalls:   1,
			wantPublishCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				deleteCalls      int
				deletedID        int64
				cacheKeyCalls    int
				cachePatternCall int
				deletedKeys      []string
				invalidatedGlob  string
				publishCalls     int
				publishedSubject string
				publishedData    []byte
			)

			repo := &mockPostRepo{
				getByIDFn: func(_ context.Context, id int64) (*domain.Post, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return &domain.Post{AuthorID: tt.postAuthorID}, nil
				},
				deleteFn: func(_ context.Context, id int64) error {
					deleteCalls++
					deletedID = id
					return tt.deleteErr
				},
			}
			cache := &mockPostCache{
				deleteFn: func(_ context.Context, keys ...string) error {
					cacheKeyCalls++
					deletedKeys = keys
					return tt.cacheErr
				},
				deleteByPatternFn: func(_ context.Context, pattern string) error {
					cachePatternCall++
					invalidatedGlob = pattern
					return tt.cacheErr
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

			svc := NewPostService(repo, cache, &mockRateLimiter{}, &mockPollProvider{}, publisher)

			err := svc.Delete(testutil.NewContext(), postID, tt.userID)

			if deleteCalls != tt.wantDeleteCalls {
				t.Errorf("postRepo.Delete calls = %d, want %d", deleteCalls, tt.wantDeleteCalls)
			}
			if cacheKeyCalls != tt.wantCacheCalls || cachePatternCall != tt.wantCacheCalls {
				t.Errorf("cache calls: Delete=%d DeleteByPattern=%d, want %d each",
					cacheKeyCalls, cachePatternCall, tt.wantCacheCalls)
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

			if deletedID != postID {
				t.Errorf("deleted post id = %d, want %d", deletedID, postID)
			}

			if len(deletedKeys) != 1 || deletedKeys[0] != "post:10" {
				t.Errorf("deleted cache keys = %v, want [post:10]", deletedKeys)
			}
			if invalidatedGlob != "feed:latest:*" {
				t.Errorf("invalidated pattern = %q, want %q", invalidatedGlob, "feed:latest:*")
			}

			if publishedSubject != core_events.SubjectPostDeleted {
				t.Errorf("subject = %q, want %q", publishedSubject, core_events.SubjectPostDeleted)
			}
			var event core_events.PostDeletedEvent
			if err := json.Unmarshal(publishedData, &event); err != nil {
				t.Fatalf("published data is not valid event JSON: %v", err)
			}
			if event.PostID != postID {
				t.Errorf("event post id = %d, want %d", event.PostID, postID)
			}
		})
	}
}
