package posts_transport_http

import (
	"net/http"

	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

// GetByID godoc
// @Summary      Get a post
// @Description  Returns a single post by ID, including its poll if present. Public endpoint.
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  PostResponse
// @Failure      400  {object}  core_http_response.ErrorBody
// @Failure      404  {object}  core_http_response.ErrorBody
// @Failure      500  {object}  core_http_response.ErrorBody
// @Router       /posts/{id} [get]
func (h *PostsHTTPHandler) GetByID(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	id, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	post, err := h.service.GetByID(ctx, int64(id))
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.JSON(http.StatusOK, NewPostResponse(post))
}
