package auth_transport_http

import (
	"context"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
)

type mockAuthService struct {
	registerFn         func(ctx context.Context, username, email, password string) (*auth_service.TokenPair, error)
	loginFn            func(ctx context.Context, email, password string) (*auth_service.TokenPair, error)
	refreshFn          func(ctx context.Context, refreshToken string) (*auth_service.TokenPair, error)
	logoutFn           func(ctx context.Context, refreshToken string) error
	getMeFn            func(ctx context.Context, userID int64) (*domain.User, error)
	parseAccessTokenFn func(tokenStr string) (int64, error)
}

func (m *mockAuthService) Register(ctx context.Context, username, email, password string) (*auth_service.TokenPair, error) {
	return m.registerFn(ctx, username, email, password)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*auth_service.TokenPair, error) {
	return m.loginFn(ctx, email, password)
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (*auth_service.TokenPair, error) {
	return m.refreshFn(ctx, refreshToken)
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return m.logoutFn(ctx, refreshToken)
}

func (m *mockAuthService) GetMe(ctx context.Context, userID int64) (*domain.User, error) {
	return m.getMeFn(ctx, userID)
}

func (m *mockAuthService) ParseAccessToken(tokenStr string) (int64, error) {
	return m.parseAccessTokenFn(tokenStr)
}
