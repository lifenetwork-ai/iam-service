package response

type Response struct {
	Status   int         `json:"status"`
	Code     string      `json:"code,omitempty"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	IsCached bool        `json:"is_cached,omitempty"`
}

type SuccessResponse struct {
	Status  int         `json:"status"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  int           `json:"status"`
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Errors  []interface{} `json:"errors,omitempty"`
}
