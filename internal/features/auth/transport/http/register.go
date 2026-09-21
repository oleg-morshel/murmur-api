package auth_transport_http

import (
	"errors"
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64,alphanum"`
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (h *AuthHTTPHandler) Register(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var req RegisterRequest

	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "invalid request body")
		return
	}

	pair, err := h.service.Register(ctx, req.Username, req.Email, req.Password)
	if errors.Is(err, core_errors.ErrConflict) {
		responseHandler.ErrorResponse(core_errors.ErrConflict, "user already exists")
		return
	}
	if err != nil {
		responseHandler.ErrorResponse(err, "internal error")
		return
	}

	responseHandler.JSON(http.StatusCreated, pair)
}
