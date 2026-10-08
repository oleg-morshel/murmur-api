package core_http_middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

const (
	requestIDKey = "X-Request-ID"
)

func RequestID() Middleware {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDKey)

			if requestID == "" {
				requestID = uuid.NewString()
			}

			r.Header.Set(requestIDKey, requestID)
			w.Header().Set(requestIDKey, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDKey)

			l := log.With(
				slog.String("request_id", requestID),
				slog.String("url", r.URL.String()),
			)

			ctx := logger.WithContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(p, "during handle HTTP request got unexpected panic")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.FromContext(ctx)
			rw := core_http_response.NewResponseWriter(w)

			before := time.Now()
			log.Debug(
				">>> incoming HTTP request",
				slog.String("http_method", r.Method),
				slog.Time("time", before.UTC()),
			)

			next.ServeHTTP(rw, r)

			log.Debug(
				"<<< done HTTP request",
				slog.Int("status_code", rw.GetStatusCode()),
				slog.Duration("latency", time.Since(before)),
			)
		})
	}
}
