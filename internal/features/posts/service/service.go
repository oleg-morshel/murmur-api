package posts_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id int64) error
}

type PostService struct {
	postRepo PostRepository
}

func NewPostService(postRepo PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}
