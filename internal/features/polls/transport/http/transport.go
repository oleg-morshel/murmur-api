package polls_transport_http

import (
	"context"
	"net/http"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_http_server "github.com/oleg-morshel/murmur-api/internal/core/transport/http/server"
)

type PollsService interface {
	CreatePoll(ctx context.Context, postID int64, question string, options []string) (*domain.Poll, error)
	Vote(ctx context.Context, pollID, optionID, userID int64) error
}

type PollsHTTPHandler struct {
	service PollsService
}

func NewPollsHTTPHandler(service PollsService) *PollsHTTPHandler {
	return &PollsHTTPHandler{service: service}
}

func (h *PollsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/posts/{id}/poll",
			Handler: h.CreatePoll,
			Public:  false,
		},
		{
			Method:  http.MethodPost,
			Path:    "/polls/{id}/vote",
			Handler: h.Vote,
			Public:  false,
		},
	}
}
