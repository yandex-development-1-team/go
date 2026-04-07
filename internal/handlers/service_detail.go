package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yandex-development-1-team/go/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// CallbackInfoPrefix prefix for this handler
const CallbackInfoPrefix = "info"

// DetailHandler handles the display of detailed information about the service
type DetailHandler struct {
	service  *botService.DetailService
	bot      *tgbotapi.BotAPI
	keyboard *KeyboardService
	sh       *StartHandler
	bs       *BoxSolutionsHandler
}

// NewDetailHandler creates a new instance of the 'DetailHandler'
func NewDetailHandler(service *botService.DetailService, bot *tgbotapi.BotAPI, sh *StartHandler, bs *BoxSolutionsHandler, keyboard *KeyboardService) *DetailHandler {
	return &DetailHandler{
		service:  service,
		bot:      bot,
		keyboard: keyboard,
		sh:       sh,
		bs:       bs,
	}
}

// Handle processes the service selection
func (h *DetailHandler) Handle(ctx context.Context, tg *tgbotapi.CallbackQuery) error {
	userID := tg.From.ID
	chatID := tg.Message.Chat.ID
	callbackData := tg.Data

	parts := strings.Split(callbackData, ":")
	back, err := h.checkBack(ctx, parts, tg)
	if err != nil {
		h.sendError(chatID, "Ошибка перехода в Главное меню")
		return err
	}
	if back {
		return nil
	}

	serviceID, err := h.service.ParseServiceID(callbackData)
	if err != nil {
		h.sendError(chatID, "Неверный формат данных")
		return err
	}

	logger.Info("service_detail_requested", zap.Int64("service_id", serviceID), zap.Int64("user_id", userID))

	service, err := h.service.GetByID(ctx, serviceID)
	if err != nil {
		return h.handleError(chatID, userID, serviceID, err)
	}

	serviceName := h.service.GetDisplayName(service)
	messageText := h.buildServiceMessage(service, serviceName)
	keyboard := h.keyboard.ServiceDetailKeyboard(service.ID)
	if err := h.sendMessage(chatID, messageText, keyboard); err != nil {
		logger.Error("failed_to_send_service_detail",
			zap.Int64("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send service detail: %w", err)
	}

	logger.Info("service_detail_shown", zap.Int64("service_id", serviceID), zap.Int64("user_id", userID))
	return nil
}

// checkBack checks for pressing the Back button
func (h *DetailHandler) checkBack(ctx context.Context, parts []string, query *tgbotapi.CallbackQuery) (bool, error) {
	if len(parts) == 2 {
		if parts[1] == "back" {
			return true, h.bs.Handle(ctx, query)
		}
	}

	return false, nil
}

// sendError sends an error message
func (h *DetailHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %s", errorMsg))
	_, err := h.bot.Send(msg)
	return err
}

// sendMessage sends a message with an inline keyboard to the user
func (h *DetailHandler) sendMessage(userID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

// handleError logs service retrieval failures and sends a message
func (h *DetailHandler) handleError(chatID int64, userID int64, serviceID int64, err error) error {
	if err != nil {
		logger.Error("failed_to_get_service",
			zap.Int64("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		h.sendError(chatID, "Не удалось загрузить информацию об услуге")
	}
	return err
}

func formatSchedule(slots []models.BoxAvailableSlot) string {
	if len(slots) == 0 {
		return ""
	}

	var sb strings.Builder

	slotsByDate := make(map[string][]models.BoxAvailableSlot)
	for _, slot := range slots {
		slotsByDate[slot.Date] = append(slotsByDate[slot.Date], slot)
	}

	for date, dateSlots := range slotsByDate {
		formattedDate := formatDateForDisplay(date)
		sb.WriteString(fmt.Sprintf("\n  %s:", formattedDate))

		for _, slot := range dateSlots {
			sb.WriteString(fmt.Sprintf("\n    • %s–%s", slot.StartTime, slot.EndTime))
		}
	}

	return sb.String()
}

// formatDateForDisplay formats the date
func formatDateForDisplay(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}

	russianMonths := []string{
		"янв", "фев", "мар", "апр", "май", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}
	russianWeekdays := []string{
		"вс", "пн", "вт", "ср", "чт", "пт", "сб",
	}

	monthIdx := date.Month() - 1
	weekdayIdx := date.Weekday()

	return fmt.Sprintf("%d %s (%s)",
		date.Day(),
		russianMonths[monthIdx],
		russianWeekdays[weekdayIdx])
}

// buildServiceMessage creates a formatted string containing all service information
func (h *DetailHandler) buildServiceMessage(service *models.Service, serviceName string) string {
	var builder strings.Builder
	builder.WriteString(serviceName)
	builder.WriteString("\n\n")

	sections := []struct {
		label     string
		value     string
		skipEmpty bool
	}{
		{"Описание", service.Description, true},
		{"Правила", service.Rules, true},
		{"Расписание", formatSchedule(service.BoxAvailableSlots), false},
	}

	for i, section := range sections {
		if section.skipEmpty && section.value == "" {
			continue
		}

		builder.WriteString(section.label)
		builder.WriteString(":")

		if section.label == "Расписание" {
			builder.WriteString(section.value)
		} else {
			builder.WriteString(" ")
			builder.WriteString(section.value)
		}
		builder.WriteString("\n")

		if i == len(sections)-1 {
			builder.WriteString("\n")
		}
	}

	builder.WriteString("Доступные действия:")
	return builder.String()
}
