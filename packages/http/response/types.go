package response

type Response struct {
	Status   int         `json:"status"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	IsCached bool        `json:"is_cached,omitempty"`
}

type ErrorResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Errors  interface{} `json:"errors"`
}

type GeneralError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

type ErrorMap struct {
	Errors map[string]interface{} `json:"errors"`
}
