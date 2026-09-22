package posts_transport_http

import (
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type UpdatePostRequest struct {
	Content   string `json:"content" validate:"required,min=1,max=5000"`
	Anonymous bool   `json:"anonymous"`
}

func (h *PostsHTTPHandler) Update(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, ok := auth_transport_http.UserIDFromContext(ctx)
	if !ok {
		resp.ErrorResponse(core_errors.ErrUnauthorized, "")
		return
	}

	id, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	var req UpdatePostRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	post, err := h.service.Update(ctx, int64(id), userID, req.Content, req.Anonymous)
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.JSON(http.StatusOK, NewPostResponse(post))
}
