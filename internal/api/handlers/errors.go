package handlers

// ErrorResponse and it's components
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorObject struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}
