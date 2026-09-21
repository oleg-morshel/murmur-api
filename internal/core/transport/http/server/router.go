package core_http_server

import (
	"fmt"
	"net/http"
)

type ApiVersion string

var (
	ApiVersion1 = ApiVersion("v1")
	ApiVersion2 = ApiVersion("v2")
	ApiVersion3 = ApiVersion("v3")
)

type APIVersionRouter struct {
	*http.ServeMux
	apiVersion     ApiVersion
	authMiddleware func(http.Handler) http.Handler
}

func NewAPIVersionRouter(apiVersion ApiVersion, authMiddleware func(http.Handler) http.Handler) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:       http.NewServeMux(),
		apiVersion:     apiVersion,
		authMiddleware: authMiddleware,
	}
}

func (r *APIVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		var handler http.Handler = route.Handler

		if !route.Public && r.authMiddleware != nil {
			handler = r.authMiddleware(handler)
		}

		r.Handle(pattern, handler)
	}
}
