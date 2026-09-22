package posts_transport_http

import (
	"context"
	"net/http"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_http_server "github.com/oleg-morshel/murmur-api/internal/core/transport/http/server"
)

type PostsService interface {
	Create(ctx context.Context, authorID int64, content string, anonymous bool) (*domain.Post, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	Update(ctx context.Context, postID, userID int64, content string, anonymous bool) (*domain.Post, error)
	Delete(ctx context.Context, id, authorID int64) error
}

type PostsHTTPHandler struct {
	service PostsService
}

func NewPostsHTTPHandler(service PostsService) *PostsHTTPHandler {
	return &PostsHTTPHandler{service: service}
}

func (h *PostsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{Method: http.MethodPost, Path: "/posts", Handler: h.Create},
		{Method: http.MethodGet, Path: "/posts", Handler: h.List, Public: true},
		{Method: http.MethodGet, Path: "/posts/{id}", Handler: h.GetByID, Public: true},
		{Method: http.MethodPut, Path: "/posts/{id}", Handler: h.Update},
		{Method: http.MethodDelete, Path: "/posts/{id}", Handler: h.Delete},
	}
}
