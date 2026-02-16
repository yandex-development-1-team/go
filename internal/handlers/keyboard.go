package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	bookHandler    = "book"
	privateButtons = "private"
	publicButtons  = "public"
	backButtons    = "no"
)

// KeyboardService генерирует клавиатуры для разных экранов
type KeyboardService struct{}

func NewKeyboardService() *KeyboardService {
	return &KeyboardService{}
}

// ServiceDetailKeyboard создаёт клавиатуру для детального просмотра услуги
func (ks *KeyboardService) ServiceDetailKeyboard(serviceType ServiceType, serviceID, boxID int) tgbotapi.InlineKeyboardMarkup {
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
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать сейчас", fmt.Sprintf("%s:%s:%d", bookHandler, backButtons, serviceID)),
			},
		}
	default:
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать", fmt.Sprintf("%s:%s:%d", bookHandler, backButtons, serviceID)),
			},
		}
	}

	// Кнопка "Назад" всегда в отдельной строке

	backButton := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("%s:%s:%d", bookHandler, backButtons, boxID))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
