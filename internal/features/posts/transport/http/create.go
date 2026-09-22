package posts_transport_http

import (
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type CreatePostRequest struct {
	Content   string `json:"content" validate:"required,min=1,max=5000"`
	Anonymous bool   `json:"anonymous"`
}

func (h *PostsHTTPHandler) Create(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, ok := auth_transport_http.UserIDFromContext(ctx)
	if !ok {
		resp.ErrorResponse(core_errors.ErrUnauthorized, "unauthorized")
		return
	}

	var req CreatePostRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		resp.ErrorResponse(err, "invalid request body")
		return
	}

	post, err := h.service.Create(ctx, userID, req.Content, req.Anonymous)
	if err != nil {
		resp.ErrorResponse(err, "internal error")
		return
	}

	resp.JSON(http.StatusCreated, NewPostResponse(post))
}
