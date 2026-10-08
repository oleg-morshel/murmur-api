package core_http_response

type ErrorBody struct {
	Message string `json:"message" example:"invalid request body"`
	Error   string `json:"error" example:"bad request"`
}
