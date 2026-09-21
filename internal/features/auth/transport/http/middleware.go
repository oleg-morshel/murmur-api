package auth_transport_http

import (
	"context"
	"net/http"
	"strings"

	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	core_http_response "github.com/oleg-morshel/murmur-api/internal/core/transport/http/response"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type TokenParser interface {
	ParseAccessToken(tokenStr string) (int64, error)
}

type ctxKey struct{}

var userIDKey = ctxKey{}

const bearerPrefix = "Bearer "

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func AuthMiddleware(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			resp := core_http_response.NewHTTPResponseHandler(log, w)

			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < len(bearerPrefix) || !strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
				resp.ErrorResponse(core_errors.ErrUnauthorized, "invalid authorization format")
				return
			}

			tokenStr := authHeader[len(bearerPrefix):]

			userID, err := parser.ParseAccessToken(tokenStr)
			if err != nil {
				resp.ErrorResponse(core_errors.ErrUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
