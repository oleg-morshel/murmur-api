package polls_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

func (s *PollService) GetByPostID(ctx context.Context, postID int64) (*domain.Poll, error) {
	return s.pollRepo.GetByPostID(ctx, postID)
}
