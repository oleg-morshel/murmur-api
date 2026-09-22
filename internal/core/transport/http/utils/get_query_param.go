package core_http_utils

import (
	"net/http"
	"strconv"
)

func GetIntQueryParam(r *http.Request, key string, defaultValue int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return val
}
