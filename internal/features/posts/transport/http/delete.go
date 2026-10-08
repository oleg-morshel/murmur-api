package posts_transport_http

import (
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

// Delete godoc
// @Summary      Delete a post
// @Description  Deletes a post. Only the author can do this.
// @Tags         posts
// @Security     BearerAuth
// @Param        id  path  int  true  "Post ID"
// @Success      204
// @Failure      400  {object}  core_http_response.ErrorBody
// @Failure      401  {object}  core_http_response.ErrorBody
// @Failure      403  {object}  core_http_response.ErrorBody  "Not the author"
// @Failure      404  {object}  core_http_response.ErrorBody
// @Failure      500  {object}  core_http_response.ErrorBody
// @Router       /posts/{id} [delete]
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
