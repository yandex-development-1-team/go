package apierrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yandex-development-1-team/go/internal/models"
)

// ServiceErrorResponse — единый формат тела ответа при ошибке (список сообщений).
type ServiceErrorResponse struct {
	Errors []string `json:"errors"`
}

// HTTPStatus возвращает HTTP-код для доменной ошибки. Неизвестные ошибки → 500.
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errors.Is(err, models.ErrUnauthorized):
		return http.StatusUnauthorized // 401
	case errors.Is(err, models.ErrForbidden):
		return http.StatusForbidden // 403
	case errors.Is(err, models.ErrUserNotFound), errors.Is(err, models.ErrBookingNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, models.ErrSlotOccupied):
		return http.StatusConflict // 409
	case errors.Is(err, models.ErrInvalidInput):
		return http.StatusBadRequest // 400
	case errors.Is(err, models.ErrRequestCanceled):
		return 499 // Client Closed Request
	case errors.Is(err, models.ErrRequestTimeout):
		return http.StatusGatewayTimeout // 504
	case errors.Is(err, models.ErrDatabase), errors.Is(err, models.ErrCache):
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError
	}
}

// Messages возвращает список человекочитаемых сообщений для ответа. SQL и внутренние детали не включаются.
func Messages(err error) []string {
	if err == nil {
		return nil
	}
	msg := messageFor(err)
	if msg == "" {
		msg = "Произошла ошибка. Попробуйте позже."
	}
	return []string{msg}
}

func messageFor(err error) string {
	switch {
	case errors.Is(err, models.ErrUnauthorized):
		return "Требуется авторизация"
	case errors.Is(err, models.ErrForbidden):
		return "Недостаточно прав"
	case errors.Is(err, models.ErrUserNotFound):
		return "Пользователь не найден"
	case errors.Is(err, models.ErrBookingNotFound):
		return "Заявка не найдена"
	case errors.Is(err, models.ErrSlotOccupied):
		return "Выбранный слот уже занят"
	case errors.Is(err, models.ErrInvalidInput):
		return "Некорректные данные"
	case errors.Is(err, models.ErrRequestCanceled):
		return "Запрос отменён"
	case errors.Is(err, models.ErrRequestTimeout):
		return "Превышено время ожидания"
	case errors.Is(err, models.ErrDatabase), errors.Is(err, models.ErrCache):
		return "Внутренняя ошибка сервиса"
	default:
		return ""
	}
}

// WriteError записывает в w ответ с единым форматом { "errors": ["..."] } и корректным статус-кодом.
// Используется для любых ошибок сервиса; детали БД/SQL не возвращаются.
func WriteError(w http.ResponseWriter, err error) {
	code := HTTPStatus(err)
	messages := Messages(err)
	WriteErrorMessages(w, code, messages)
}

// WriteErrorMessages записывает в w ответ с заданным кодом и списком сообщений.
func WriteErrorMessages(w http.ResponseWriter, code int, messages []string) {
	if len(messages) == 0 {
		messages = []string{"Произошла ошибка. Попробуйте позже."}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ServiceErrorResponse{Errors: messages})
}
