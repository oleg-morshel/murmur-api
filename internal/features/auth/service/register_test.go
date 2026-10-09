package auth_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	errDB := errors.New("db down")

	const (
		username  = "oleg"
		password  = "super-secret-password"
		wantEmail = "oleg@example.com"
		userID    = int64(42)
	)

	tests := []struct {
		name          string
		email         string
		createErr     error
		saveErr       error
		wantErr       error
		wantSaveCalls int
	}{
		{
			name:          "success",
			email:         "oleg@example.com",
			wantSaveCalls: 1,
		},
		{
			name:          "email is trimmed and lowercased",
			email:         "  Oleg@Example.COM ",
			wantSaveCalls: 1,
		},
		{
			name:          "duplicate email",
			email:         "oleg@example.com",
			createErr:     core_errors.ErrConflict,
			wantErr:       core_errors.ErrConflict,
			wantSaveCalls: 0,
		},
		{
			name:          "repository error",
			email:         "oleg@example.com",
			createErr:     errDB,
			wantErr:       errDB,
			wantSaveCalls: 0,
		},
		{
			name:          "refresh token save fails",
			email:         "oleg@example.com",
			saveErr:       errDB,
			wantErr:       errDB,
			wantSaveCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotUser   *domain.User
				gotToken  *domain.RefreshToken
				saveCalls int
			)

			users := &mockUserRepo{
				createFn: func(_ context.Context, u *domain.User) (int64, error) {
					gotUser = u
					if tt.createErr != nil {
						return 0, tt.createErr
					}
					return userID, nil
				},
			}
			tokens := &mockTokenRepo{
				saveFn: func(_ context.Context, tk *domain.RefreshToken) error {
					saveCalls++
					gotToken = tk
					return tt.saveErr
				},
			}

			svc := NewService(users, tokens, "test-secret", 15*time.Minute, 24*time.Hour)

			pair, err := svc.Register(context.Background(), username, tt.email, password)

			if saveCalls != tt.wantSaveCalls {
				t.Errorf("tokenRepo.Save calls = %d, want %d", saveCalls, tt.wantSaveCalls)
			}

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if pair != nil {
					t.Errorf("tokens = %+v, want nil on error", pair)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pair == nil || pair.AccessToken == "" || pair.RefreshToken == "" {
				t.Fatalf("expected non-empty token pair, got %+v", pair)
			}

			if gotUser.Email != wantEmail {
				t.Errorf("email = %q, want %q", gotUser.Email, wantEmail)
			}
			if gotUser.Username != username {
				t.Errorf("username = %q, want %q", gotUser.Username, username)
			}

			if gotUser.PasswordHash == password {
				t.Fatal("password stored in plain text")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(gotUser.PasswordHash), []byte(password)); err != nil {
				t.Errorf("password hash does not match password: %v", err)
			}

			if gotToken.UserID != userID {
				t.Errorf("refresh token user id = %d, want %d", gotToken.UserID, userID)
			}

			gotID, err := svc.ParseAccessToken(pair.AccessToken)
			if err != nil {
				t.Fatalf("access token does not parse: %v", err)
			}
			if gotID != userID {
				t.Errorf("access token user id = %d, want %d", gotID, userID)
			}
		})
	}
}
