package polls_transport_http

import (
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type VoteRequest struct {
	OptionID int64 `json:"option_id" validate:"required"`
}

func (h *PollsHTTPHandler) Vote(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	pollID, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	var req VoteRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	userID, ok := auth_transport_http.UserIDFromContext(ctx)
	if !ok {
		resp.ErrorResponse(core_errors.ErrUnauthorized, "")
		return
	}

	if err := h.service.Vote(ctx, int64(pollID), req.OptionID, userID); err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.NoContentResponse()
}
