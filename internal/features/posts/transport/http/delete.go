package posts_transport_http

import (
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

func (h *PostsHTTPHandler) Delete(rw http.ResponseWriter, r *http.Request) {
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

	if err := h.service.Delete(ctx, int64(id), userID); err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.NoContentResponse()
}
