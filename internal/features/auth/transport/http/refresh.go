package auth_transport_http

import (
	"errors"
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Exchanges a valid refresh token for a new access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RefreshRequest  true  "Refresh token"
// @Success      200      {object}  auth_service.TokenPair
// @Failure      400      {object}  core_http_response.ErrorBody
// @Failure      401      {object}  core_http_response.ErrorBody  "Invalid or expired token"
// @Failure      500      {object}  core_http_response.ErrorBody
// @Router       /auth/refresh [post]
func (h *AuthHTTPHandler) Refresh(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var req RefreshRequest

	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	pair, err := h.service.Refresh(ctx, req.RefreshToken)

	if errors.Is(err, auth_service.ErrInvalidToken) {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "invalid or expired token")
		return
	}

	if err != nil {
		responseHandler.ErrorResponse(err, "internal server error")
		return
	}

	responseHandler.JSON(http.StatusOK, pair)
}
