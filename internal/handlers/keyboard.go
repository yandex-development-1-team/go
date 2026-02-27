package handlers

import (
	"fmt"
	"time"

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
	backButton := tgbotapi.NewInlineKeyboardButtonData(BackButtonsTitle, fmt.Sprintf("%s:%d", backButtons, boxID))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// FormNavigationKeyboard creates a keyboard with navigation for the steps of the form
func (ks *KeyboardService) FormNavigationKeyboard() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Назад", "book:main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// ConfirmationKeyboard creates a keyboard for the confirmation step
func (ks *KeyboardService) ConfirmationKeyboard() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "book:confirm"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("Отменить", "book:main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// createNoDatesKeyboard creates a 'To Main Menu' button
func (ks *KeyboardService) createMainMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "book:main_menu"),
		),
	)
	return &keyboard
}

// createDatesKeyboard creates an inline keyboard with available dates
func (ks *KeyboardService) createDatesKeyboard(dates []time.Time) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Grouping the dates by 2 in a row
	for i := 0; i < len(dates); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// The first button in the row
		btn1 := tgbotapi.NewInlineKeyboardButtonData(
			ks.formatDateButton(dates[i]),
			fmt.Sprintf("book:select_date:%s", dates[i].Format("2006-01-02")),
		)
		row = append(row, btn1)

		// The second button, if available
		if i+1 < len(dates) {
			btn2 := tgbotapi.NewInlineKeyboardButtonData(
				ks.formatDateButton(dates[i+1]),
				fmt.Sprintf("book:select_date:%s", dates[i+1].Format("2006-01-02")),
			)
			row = append(row, btn2)
		}

		rows = append(rows, row)
	}

	// Adding the Back button
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад", "book:main_menu"),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// formatDateButton formats the date to display on the button
func (ks *KeyboardService) formatDateButton(date time.Time) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if date.YearDay() == today.YearDay() && date.Year() == today.Year() {
		return "Сегодня"
	}

	tomorrow := today.AddDate(0, 0, 1)
	if date.YearDay() == tomorrow.YearDay() && date.Year() == tomorrow.Year() {
		return "Завтра"
	}

	// Format: "Jan 15 (Mon)"
	return date.Format("02 Jan (Mon)")
}
