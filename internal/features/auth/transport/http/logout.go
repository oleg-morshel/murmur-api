package auth_transport_http

import (
	"net/http"

	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Logout godoc
// @Summary      Log out
// @Description  Invalidates the given refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  LogoutRequest  true  "Refresh token"
// @Success      204
// @Failure      400  {object}  core_http_response.ErrorBody
// @Failure      500  {object}  core_http_response.ErrorBody
// @Router       /auth/logout [post]
func (h *AuthHTTPHandler) Logout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var req LogoutRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	if err := h.service.Logout(ctx, req.RefreshToken); err != nil {
		responseHandler.ErrorResponse(err, "logout failed")
		return
	}
	responseHandler.NoContentResponse()
}
