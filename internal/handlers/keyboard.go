package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ServiceType определяет тип услуги для выбора правильного набора кнопок
type ServiceType string

const (
	ServiceTypeMuseum  ServiceType = "museum"
	ServiceTypeSport   ServiceType = "sport"
	ServiceTypeDefault ServiceType = "default"
)

// используется для определения значения serviceType
const (
	bookHandler      = "book"
	privateButtons   = "private"
	publicButtons    = "public"
	backButtons      = "back"
	BackButtonsTitle = "⬅️ Назад"
	missingParameter = "no"
)

// KeyboardService генерирует клавиатуры для разных экранов
type KeyboardService struct{}

func NewKeyboardService() *KeyboardService {
	return &KeyboardService{}
}

// ServiceDetailKeyboard создаёт клавиатуру для детального просмотра услуги
func (ks *KeyboardService) ServiceDetailKeyboard(serviceType ServiceType, serviceID int64) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	switch serviceType {
	case ServiceTypeMuseum:
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("👤 Приватный тур", fmt.Sprintf("%s:%s:%d", bookHandler, privateButtons, serviceID)),
				tgbotapi.NewInlineKeyboardButtonData("👥 Групповой тур", fmt.Sprintf("%s:%s:%d", bookHandler, publicButtons, serviceID)),
			},
		}
	case ServiceTypeSport:
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать сейчас", fmt.Sprintf("%s:%s:%d", bookHandler, missingParameter, serviceID)),
			},
		}
	default:
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать", fmt.Sprintf("%s:%s:%d", bookHandler, missingParameter, serviceID)),
			},
		}
	}

	// Кнопка "Назад" всегда в отдельной строке
	backButton := tgbotapi.NewInlineKeyboardButtonData(BackButtonsTitle, fmt.Sprintf("%s:%d", backButtons, serviceID))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func getBackButton(alias string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData("Назад", alias)
}
