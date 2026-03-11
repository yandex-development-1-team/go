package handlers

<<<<<<< HEAD
// ErrorResponse and it's components
=======
// Message-style response (auth, legacy)
type MessageStatus string

const (
	StatusError   MessageStatus = "error"
	StatusSuccess MessageStatus = "success"
	StatusInfo    MessageStatus = "info"
)

type MessageDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string          `json:"message"`
	Status  MessageStatus   `json:"status"`
	Details []MessageDetail `json:"details,omitempty"`
}

// Error-style response (REST API, e.g. special_project)
>>>>>>> dev
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
<<<<<<< HEAD
=======

// Helpers: MessageResponse for auth flow
func UnauthorizedError(detail string) *MessageResponse {
	return &MessageResponse{
		Message: "Неавторизован",
		Status:  StatusError,
		Details: []MessageDetail{{Message: detail}},
	}
}

func BadRequestError() *MessageResponse {
	return &MessageResponse{
		Message: "Не валидные входные данные",
		Status:  StatusError,
	}
}

func ForbiddenError() *MessageResponse {
	return &MessageResponse{
		Message: "Недостаточно прав",
		Status:  StatusError,
	}
}

func InternalServerError() *MessageResponse {
	return &MessageResponse{
		Message: "Внутренняя ошибка сервера",
		Status:  StatusError,
	}
}
>>>>>>> dev
