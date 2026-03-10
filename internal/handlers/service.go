package handlers

import (
	"errors"
	"strconv"
	"strings"

	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/yandex-development-1-team/go/internal/models"
)

var (
	ErrNullData      = errors.New("handler with empty data")
	ErrIncorrectData = errors.New("handler with incorrect data")
	ErrInvalidField  = errors.New("handler with an invalid field")
)

type ServiceRepo interface {
	GetServiceByID(ctx context.Context, serviceID int) (models.Service, error)
}

type Service struct {
	ID             int64
	Name           string
	Description    string
	Rules          string
	Schedule       string
	AvailableSlots []AvailableSlot
	Type           ServiceType
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}

type ServiceHandler struct {
	repo            ServiceRepo
	bot             *tgbotapi.BotAPI
	keyboardService *KeyboardService
}

func (h *ServiceHandler) sendMessage(userID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

func NewServiceHandler(repo ServiceRepo, bot *tgbotapi.BotAPI) *ServiceHandler {
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
