package apierrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/models"
)

type ServiceErrorResponse struct {
	Errors []string `json:"errors"`
}

type errMapping struct {
	err     error
	status  int
	message string
}

var errMappings = []errMapping{
	{models.ErrUnauthorized, http.StatusUnauthorized, "Требуется авторизация"},
	{models.ErrForbidden, http.StatusForbidden, "Недостаточно прав"},
	{models.ErrUserBlocked, http.StatusForbidden, "Учётная запись заблокирована"},
	{models.ErrInvalidCredentials, http.StatusBadRequest, "Неверный логин или пароль"},
	{models.ErrUserNotFound, http.StatusNotFound, "Пользователь не найден"},
	{models.ErrBookingNotFound, http.StatusNotFound, "Заявка не найдена"},
	{models.ErrSpecialProjectNotFound, http.StatusNotFound, "Спецпроект не найден"},
	{models.ErrApplicationNotFound, http.StatusNotFound, "Заявка на спец проект не найдена"},
	{models.ErrSlotOccupied, http.StatusConflict, "Выбранный слот уже занят"},
	{models.ErrInvalidInput, http.StatusBadRequest, "Некорректные данные"},
	{models.ErrRequestCanceled, 499, "Запрос отменён"},
	{models.ErrRequestTimeout, http.StatusGatewayTimeout, "Превышено время ожидания"},
	{models.ErrDatabase, http.StatusInternalServerError, "Внутренняя ошибка сервиса"},
	{models.ErrCache, http.StatusInternalServerError, "Внутренняя ошибка сервиса"},
	{models.ErrEmailAlreadyExist, http.StatusConflict, "Указанный почтовый адрес уже занят"},
	{models.ErrSlotsNotFound, http.StatusBadRequest, "Слоты коробочного решения не найдены"},
	{models.ErrBoxSolutionNotFound, http.StatusNotFound, "Коробочное решение не найдено"},
}

const defaultMessage = "Произошла ошибка. Попробуйте позже."

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	for _, m := range errMappings {
		if errors.Is(err, m.err) {
			return m.status
		}
	}
	return http.StatusInternalServerError
}

func Messages(err error) []string {
	if err == nil {
		return nil
	}
	msg := messageFor(err)
	if msg == "" {
		msg = defaultMessage
	}
	return []string{msg}
}

func messageFor(err error) string {
	for _, m := range errMappings {
		if errors.Is(err, m.err) {
			return m.message
		}
	}
	return ""
}

func WriteError(w http.ResponseWriter, err error) {
	code := HTTPStatus(err)
	messages := Messages(err)
	WriteErrorMessages(w, code, messages)
}

func WriteErrorMessages(w http.ResponseWriter, code int, messages []string) {
	if len(messages) == 0 {
		messages = []string{defaultMessage}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ServiceErrorResponse{Errors: messages})
}

func WriteErrorGin(c *gin.Context, err error) {
	c.Abort()
	WriteError(c.Writer, err)
}

func WriteErrorMessagesGin(c *gin.Context, code int, messages []string) {
	c.Abort()
	if len(messages) == 0 {
		messages = []string{defaultMessage}
	}
	WriteErrorMessages(c.Writer, code, messages)
}
