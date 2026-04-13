package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/yandex-development-1-team/go/internal/models"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
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
func (ks *KeyboardService) ServiceDetailKeyboard(serviceID int64) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	buttons = [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать", fmt.Sprintf("%s:%d", bookHandler, serviceID)),
		},
	}

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

// DatesKeyboard creates an inline keyboard with available dates and time slots
func (ks *KeyboardService) DatesKeyboard(slots []models.BoxAvailableSlot) tgbotapi.InlineKeyboardMarkup {
	if len(slots) == 0 {
		return tgbotapi.InlineKeyboardMarkup{}
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Группируем слоты по 2 в строку
	for i := 0; i < len(slots); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// Первая кнопка в строке
		slot1 := (slots)[i]
		btn1 := tgbotapi.NewInlineKeyboardButtonData(
			ks.formatSlotButton(slot1),
			ks.buildSlotCallback(slot1),
		)
		row = append(row, btn1)

		// Вторая кнопка, если есть
		if i+1 < len(slots) {
			slot2 := (slots)[i+1]
			btn2 := tgbotapi.NewInlineKeyboardButtonData(
				ks.formatSlotButton(slot2),
				ks.buildSlotCallback(slot2),
			)
			row = append(row, btn2)
		}

		rows = append(rows, row)
	}

	navKeyboard := ks.FormNavigationKeyboard(botService.StepReturnInBoxList)
	rows = append(rows, navKeyboard.InlineKeyboard...)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildSlotCallback generates callback data for the slot
func (ks *KeyboardService) buildSlotCallback(slot models.BoxAvailableSlot) string {
	startTime := strings.ReplaceAll(slot.StartTime, ":", ".")
	endTime := strings.ReplaceAll(slot.EndTime, ":", ".")
	return fmt.Sprintf("book:select_date:%s:%s:%s", slot.Date, startTime, endTime)
}

// formatSlotButton formats the date to display on the button
func (ks *KeyboardService) formatSlotButton(slot models.BoxAvailableSlot) string {
	date, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		return slot.Date
	}

	var dateStr string
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if date.YearDay() == today.YearDay() && date.Year() == today.Year() {
		dateStr = "Сегодня"
	} else if date.YearDay() == today.AddDate(0, 0, 1).YearDay() && date.Year() == today.Year() {
		dateStr = "Завтра"
	} else {
		russianMonths := []string{
			"янв", "фев", "мар", "апр", "май", "июн",
			"июл", "авг", "сен", "окт", "ноя", "дек",
		}
		russianWeekdays := []string{
			"вс", "пн", "вт", "ср", "чт", "пт", "сб",
		}

		monthIdx := date.Month() - 1
		weekdayIdx := date.Weekday()
		dateStr = fmt.Sprintf("%02d %s (%s)",
			date.Day(),
			russianMonths[monthIdx],
			russianWeekdays[weekdayIdx])
	}

	timeStr := fmt.Sprintf("%s-%s", slot.StartTime, slot.EndTime)

	return fmt.Sprintf("%s\n%s", dateStr, timeStr)
}

// CreateButton creates a button with 'text' and 'data'
func (ks *KeyboardService) CreateButton(text string, data string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, data),
		),
	)
	return &keyboard
}
