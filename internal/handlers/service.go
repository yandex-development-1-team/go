package handlers

import (
	"errors"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	repository "github.com/yandex-development-1-team/go/internal/repository"
)

var (
	ErrNullData      = errors.New("handler with empty data")
	ErrIncorrectData = errors.New("handler with incorrect data")
	ErrInvalidField  = errors.New("handler with an invalid field")
)

// Service представляет собой услугу с полной информацией
type Service struct {
	ID          int         //Уникальный идентификатор услуги
	Name        string      // название услуги
	Description string      // описание
	Rules       string      // правила
	Schedule    string      // время проведения
	Type        ServiceType // Тип услуги (музей, спорт и т.д.)
	BoxID       int         // ID бокса/категории для кнопки "Назад"
}

// ServiceHandler обрабатывает действия, связанные с услугами
type ServiceHandler struct {
	repo            *repository.Repository
	bot             *tgbotapi.BotAPI
	keyboardService *KeyboardService
}

func (h *ServiceHandler) sendMessage(userID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

// NewServiceHandler создаёт новый обработчик услуг
func NewServiceHandler(repo *repository.Repository, bot *tgbotapi.BotAPI) *ServiceHandler {
	return &ServiceHandler{
		repo:            repo,
		bot:             bot,
		keyboardService: NewKeyboardService(),
	}
}

// splitHanlerData проверяет полученные данные, и возвращает ID услуги
func parseServiceID(s string) (int, error) {
	if s == "" {
		return 0, ErrNullData
	}
	info := strings.Split(s, ":")
	if len(info) != 2 {
		return 0, ErrIncorrectData
	}
	if info[0] != "info" {
		return 0, ErrInvalidField
	}
	i, err := strconv.Atoi(info[1])
	if err != nil {
		return 0, err
	}
	return i, nil
}
