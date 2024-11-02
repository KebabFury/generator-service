package responses

type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []interface{} `json:"errors,omitempty"`
}

type PingResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type Response[T any] struct {
	Data []T `json:"data"`
}
