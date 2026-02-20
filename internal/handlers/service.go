package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/repository" // название пакета с запросом к ДБ
	"go.uber.org/zap"
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
	logger          *zap.Logger
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
func NewServiceHandler(logger *zap.Logger, repo *repository.Repository, bot *tgbotapi.BotAPI) *ServiceHandler {
	return &ServiceHandler{
		logger:          logger.Named("service_handler"),
		repo:            repo,
		bot:             bot,
		keyboardService: NewKeyboardService(),
	}
}
