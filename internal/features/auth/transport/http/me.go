package auth_transport_http

import (
	"errors"
	"net/http"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func userResponseFromDomain(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// GetMe godoc
// @Summary      Current user
// @Description  Returns the profile of the authenticated user.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  core_http_response.ErrorBody
// @Failure      404  {object}  core_http_response.ErrorBody
// @Failure      500  {object}  core_http_response.ErrorBody
// @Router       /auth/me [get]
func (h *AuthHTTPHandler) GetMe(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		responseHandler.ErrorResponse(core_errors.ErrUnauthorized, "")
		return
	}

	user, err := h.service.GetMe(ctx, userID)
	if errors.Is(err, core_errors.ErrNotFound) {
		responseHandler.ErrorResponse(core_errors.ErrNotFound, "user not found")
		return
	}

	if err != nil {
		responseHandler.ErrorResponse(err, "internal server error")
		return
	}

	responseHandler.JSON(http.StatusOK, userResponseFromDomain(user))
}
