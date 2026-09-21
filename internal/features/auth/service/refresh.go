package auth_service

import (
	"context"
	"fmt"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := hashToken(refreshToken)

	userID, err := s.tokenRepo.ConsumeByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("auth.Refresh: %w", err)
	}
	if userID == 0 {
		return nil, ErrInvalidToken
	}

	return s.generateTokenPair(ctx, userID)
}
