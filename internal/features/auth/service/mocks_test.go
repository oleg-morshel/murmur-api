package auth_service

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, user *domain.User) (int64, error)
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIdFn    func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) (int64, error) {
	return m.createFn(ctx, user)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) GetById(ctx context.Context, id int64) (*domain.User, error) {
	return m.getByIdFn(ctx, id)
}

type mockTokenRepo struct {
	saveFn          func(ctx context.Context, token *domain.RefreshToken) error
	deleteByHashFn  func(ctx context.Context, tokenHash string) (bool, error)
	consumeByHashFn func(ctx context.Context, tokenHash string) (int64, error)
}

func (m *mockTokenRepo) Save(ctx context.Context, token *domain.RefreshToken) error {
	return m.saveFn(ctx, token)
}

func (m *mockTokenRepo) DeleteByHash(ctx context.Context, tokenHash string) (bool, error) {
	return m.deleteByHashFn(ctx, tokenHash)
}

func (m *mockTokenRepo) ConsumeByHash(ctx context.Context, tokenHash string) (int64, error) {
	return m.consumeByHashFn(ctx, tokenHash)
}
