package auth_transport_http

import (
	"errors"
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	"github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// Login godoc
// @Summary      Log in
// @Description  Exchanges email and password for an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Credentials"
// @Success      200      {object}  auth_service.TokenPair
// @Failure      400      {object}  core_http_response.ErrorBody
// @Failure      401      {object}  core_http_response.ErrorBody  "Invalid email or password"
// @Failure      500      {object}  core_http_response.ErrorBody
// @Router       /auth/login [post]
func (h *AuthHTTPHandler) Login(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	pair, err := h.service.Login(ctx, request.Email, request.Password)
	if errors.Is(err, auth_service.ErrInvalidCredentials) {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "invalid email or password")
		return
	}

	if err != nil {
		responseHandler.ErrorResponse(err, "internal error")
		return
	}

	responseHandler.JSON(http.StatusOK, pair)
}
