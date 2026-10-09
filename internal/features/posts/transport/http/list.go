package posts_transport_http

import (
	"net/http"

	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

// List godoc
// @Summary      List posts
// @Description  Returns posts with pagination. Public endpoint.
// @Tags         posts
// @Produce      json
// @Param        limit   query     int  false  "Page size"  default(20)
// @Param        offset  query     int  false  "Offset"     default(0)
// @Success      200     {array}   PostResponse
// @Failure      400     {object}  core_http_response.ErrorBody
// @Failure      500     {object}  core_http_response.ErrorBody
// @Router       /posts [get]
func (h *PostsHTTPHandler) List(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	limit := core_http_utils.GetIntQueryParam(r, "limit", 20)
	offset := core_http_utils.GetIntQueryParam(r, "offset", 0)

	posts, err := h.service.List(ctx, limit, offset)
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.JSON(http.StatusOK, NewPostListResponse(posts))
}
