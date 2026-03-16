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
	backButton := tgbotapi.NewInlineKeyboardButtonData(BackButtonsTitle, fmt.Sprintf("info:%s", backButtons))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func getBackButton(alias string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData("Назад", alias)
}

// FormNavigationKeyboard creates a keyboard with navigation for the steps of the form
func (ks *KeyboardService) FormNavigationKeyboard(step int) tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			getBackButton(fmt.Sprintf("book:back:%d", step)),
			//tgbotapi.NewInlineKeyboardButtonData("Назад", fmt.Sprintf("back:%d", step)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "book:main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// ConfirmationKeyboard creates a keyboard for the confirmation step
func (ks *KeyboardService) ConfirmationKeyboard(step int) tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "book:confirm"),
			getBackButton("book:back:4"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("Отменить", "book:main_menu"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// MainMenuKeyboard creates a 'To Main Menu' button
func (ks *KeyboardService) MainMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "book:main_menu"),
		),
	)
	return &keyboard
}

// DatesKeyboard creates an inline keyboard with available dates
func (ks *KeyboardService) DatesKeyboard(visitType string, dates []time.Time) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Grouping the dates by 2 in a row
	for i := 0; i < len(dates); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// The first button in the row
		btn1 := tgbotapi.NewInlineKeyboardButtonData(
			ks.formatDateButton(dates[i]),
			fmt.Sprintf("book:select_date:%s:%s", visitType, dates[i].Format("2006-01-02")),
		)
		row = append(row, btn1)

		// The second button, if available
		if i+1 < len(dates) {
			btn2 := tgbotapi.NewInlineKeyboardButtonData(
				ks.formatDateButton(dates[i+1]),
				fmt.Sprintf("book:select_date:%s:%s", visitType, dates[i+1].Format("2006-01-02")),
			)
			row = append(row, btn2)
		}

		rows = append(rows, row)
	}

	// Adding the Back button
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("В главное меню", "book:main_menu"),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// formatDateButton formats the date to display on the button
func (ks *KeyboardService) formatDateButton(date time.Time) string {
	var russianMonths = []string{
		"янв", "фев", "мар", "апр", "май", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}

	// Русские названия дней недели (короткие)
	var russianWeekdays = []string{
		"вс", "пн", "вт", "ср", "чт", "пт", "сб",
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if date.YearDay() == today.YearDay() && date.Year() == today.Year() {
		return "Сегодня"
	}

	tomorrow := today.AddDate(0, 0, 1)
	if date.YearDay() == tomorrow.YearDay() && date.Year() == tomorrow.Year() {
		return "Завтра"
	}

	monthIdx := date.Month() - 1
	weekdayIdx := date.Weekday() // Воскресенье = 0 в Go

	return fmt.Sprintf("%02d %s (%s)",
		date.Day(),
		russianMonths[monthIdx],
		russianWeekdays[weekdayIdx])
}
