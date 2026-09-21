package auth_service

import (
	"context"
)

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	_, err := s.tokenRepo.DeleteByHash(ctx, hash)
	return err
}
