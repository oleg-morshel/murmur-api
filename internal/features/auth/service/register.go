package auth_service

import (
	"context"
	"fmt"
	"strings"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(ctx context.Context, username, email, password string) (*TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth.Register: hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	userID, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, userID)
}
