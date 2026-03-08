package apierrors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
)

// Пакет предназначен для использования из Gin-хендлеров и middleware.
// В хендлерах удобнее вызывать WriteErrorGin(c, err) / WriteErrorMessagesGin(c, code, messages) —
// они делают c.Abort() и пишут ответ в формате ServiceErrorResponse.
// Middleware или код без *gin.Context могут использовать WriteError(w, err), передавая w (например c.Writer).

// ServiceErrorResponse — единый формат тела ответа при ошибке (список сообщений).
type ServiceErrorResponse struct {
	Errors []string `json:"errors"`
}

// errMapping связывает доменную ошибку с HTTP-кодом и сообщением для клиента.
type errMapping struct {
	err     error
	status  int
	message string
}

// errMappings — единая таблица «ошибка → статус + сообщение». Порядок важен: первое совпадение.
var errMappings = []errMapping{
	{models.ErrUnauthorized, http.StatusUnauthorized, "Требуется авторизация"},
	{models.ErrForbidden, http.StatusForbidden, "Недостаточно прав"},
	{models.ErrUserBlocked, http.StatusForbidden, "Учётная запись заблокирована"},
	{models.ErrInvalidCredentials, http.StatusUnauthorized, "Неверный логин или пароль"},
	{models.ErrUserNotFound, http.StatusNotFound, "Пользователь не найден"},
	{models.ErrBookingNotFound, http.StatusNotFound, "Заявка не найдена"},
	{models.ErrSpecProjNotFound, http.StatusNotFound, "Спецпроект не найден"},
	{service.ErrNotFound, http.StatusNotFound, "Спецпроект не найден"},
	{models.ErrSlotOccupied, http.StatusConflict, "Выбранный слот уже занят"},
	{models.ErrInvalidInput, http.StatusBadRequest, "Некорректные данные"},
	{models.ErrRequestCanceled, 499, "Запрос отменён"},
	{models.ErrRequestTimeout, http.StatusGatewayTimeout, "Превышено время ожидания"},
	{models.ErrDatabase, http.StatusInternalServerError, "Внутренняя ошибка сервиса"},
	{models.ErrCache, http.StatusInternalServerError, "Внутренняя ошибка сервиса"},
}

const defaultMessage = "Произошла ошибка. Попробуйте позже."

// HTTPStatus возвращает HTTP-код для доменной ошибки. Неизвестные ошибки → 500.
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

// Messages возвращает список человекочитаемых сообщений для ответа.
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

// WriteError записывает в w ответ с форматом { "errors": ["..."] } и корректным статус-кодом.
func WriteError(w http.ResponseWriter, err error) {
	code := HTTPStatus(err)
	messages := Messages(err)
	WriteErrorMessages(w, code, messages)
}

// WriteErrorMessages записывает в w ответ с заданным кодом и списком сообщений.
func WriteErrorMessages(w http.ResponseWriter, code int, messages []string) {
	if len(messages) == 0 {
		messages = []string{defaultMessage}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ServiceErrorResponse{Errors: messages})
}

// WriteErrorGin прерывает цепочку Gin и отправляет ошибку в формате ServiceErrorResponse.
// Для использования в Gin-хендлерах: apierrors.WriteErrorGin(c, err); return
func WriteErrorGin(c *gin.Context, err error) {
	c.Abort()
	WriteError(c.Writer, err)
}

// WriteErrorMessagesGin прерывает цепочку Gin и отправляет ответ с заданным кодом и списком сообщений.
// Пустой messages обрабатывается так же, как в WriteErrorMessages — подставляется defaultMessage.
func WriteErrorMessagesGin(c *gin.Context, code int, messages []string) {
	c.Abort()
	if len(messages) == 0 {
		messages = []string{defaultMessage}
	}
	WriteErrorMessages(c.Writer, code, messages)
}
