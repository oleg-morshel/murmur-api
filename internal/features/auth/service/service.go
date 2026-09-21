package auth_service

import (
	"context"
	"errors"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetById(ctx context.Context, id int64) (*domain.User, error)
}

type TokenRepository interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	DeleteByHash(ctx context.Context, tokenHash string) (bool, error)
	ConsumeByHash(ctx context.Context, tokenHash string) (int64, error)
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Service struct {
	userRepo   UserRepository
	tokenRepo  TokenRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}
