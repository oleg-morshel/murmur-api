package auth_transport_http

import (
	"context"
	"net/http"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_http_server "github.com/oleg-morshel/murmur-api/internal/core/transport/http/server"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*auth_service.TokenPair, error)
	Login(ctx context.Context, email, password string) (*auth_service.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*auth_service.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	GetMe(ctx context.Context, userID int64) (*domain.User, error)
	ParseAccessToken(tokenStr string) (int64, error)
}

type AuthHTTPHandler struct {
	service AuthService
}

func NewAuthHTTPHandler(svc AuthService) *AuthHTTPHandler {
	return &AuthHTTPHandler{service: svc}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{Method: http.MethodPost, Path: "/auth/register", Public: true, Handler: h.Register},
		{Method: http.MethodPost, Path: "/auth/login", Public: true, Handler: h.Login},
		{Method: http.MethodPost, Path: "/auth/refresh", Public: true, Handler: h.Refresh},
		{Method: http.MethodPost, Path: "/auth/logout", Public: true, Handler: h.Logout},
		{Method: http.MethodGet, Path: "/auth/me", Public: false, Handler: h.GetMe},
	}
}
