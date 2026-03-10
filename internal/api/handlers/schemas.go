package handlers

type MessageStatus string

const (
	StatusError   MessageStatus = "error"
	StatusSuccess MessageStatus = "success"
	StatusInfo    MessageStatus = "info"
)

type MessageResponse struct {
	Message string          `json:"message"`
	Status  MessageStatus   `json:"status"`
	Details []MessageDetail `json:"details,omitempty"`
}

type MessageDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// type ErrorDetail struct {
// 	Field   string `json:"field"`
// 	Message string `json:"message"`
// }

// type ErrorBody struct {
// 	Code    string        `json:"code"`
// 	Message string        `json:"message"`
// 	Details []ErrorDetail `json:"details,omitempty"`
// }

// type ErrorResponse struct {
// 	Error ErrorBody `json:"error"`
// }
