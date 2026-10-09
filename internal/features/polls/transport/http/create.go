package polls_transport_http

import (
	"net/http"

	core_http_request "github.com/oleg-morshel/murmur-api/internal/core/transport/http/request"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	core_http_utils "github.com/oleg-morshel/murmur-api/internal/core/transport/http/utils"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type CreatePollRequest struct {
	Question string   `json:"question" validate:"required,min=1,max=500"`
	Options  []string `json:"options" validate:"required,min=2,max=10,dive,required,min=1,max=200"`
}

// CreatePoll godoc
// @Summary      Create a poll for a post
// @Description  Attaches a poll (2–10 options) to an existing post.
// @Tags         polls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                true  "Post ID"
// @Param        request  body      CreatePollRequest  true  "Poll data"
// @Success      201      {object}  PollResponse
// @Failure      400      {object}  core_http_response.ErrorBody
// @Failure      401      {object}  core_http_response.ErrorBody
// @Failure      404      {object}  core_http_response.ErrorBody
// @Failure      409      {object}  core_http_response.ErrorBody  "Post already has a poll"
// @Failure      500      {object}  core_http_response.ErrorBody
// @Router       /posts/{id}/poll [post]
func (h *PollsHTTPHandler) CreatePoll(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	resp := core_http_response.NewHTTPResponseHandler(log, rw)

	postID, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	var req CreatePollRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	poll, err := h.service.CreatePoll(ctx, int64(postID), req.Question, req.Options)
	if err != nil {
		resp.ErrorResponse(err, "")
		return
	}

	resp.JSON(http.StatusCreated, NewPollResponse(poll))
}
