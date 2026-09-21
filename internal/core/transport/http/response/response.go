package core_http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type HTTPResponseHandler struct {
	log *logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSON(code int, data interface{}) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(code)
	if err := json.NewEncoder(h.rw).Encode(data); err != nil {
		h.log.Error("failed to write response", slog.Any("error", err))
	}
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		code    int
		logFunc func(string, ...any)
	)

	switch {
	case errors.Is(err, core_errors.ErrBadRequest):
		code = http.StatusBadRequest
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUnauthorized):
		code = http.StatusUnauthorized
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrForbidden):
		code = http.StatusForbidden
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrNotFound):
		code = http.StatusNotFound
		logFunc = h.log.Debug
	case errors.Is(err, core_errors.ErrConflict):
		code = http.StatusConflict
		logFunc = h.log.Warn
	default:
		code = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, slog.Any("error", err))
	h.errorResponse(code, err, msg)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError

	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, slog.Any("error", err))
	h.errorResponse(statusCode, err, msg)
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) errorResponse(statusCode int, err error, msg string) {
	errText := err.Error()
	if statusCode == http.StatusInternalServerError {
		errText = "internal server error"
	}

	response := map[string]string{
		"message": msg,
		"error":   errText,
	}
	h.JSON(statusCode, response)
}
