package handlers

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
