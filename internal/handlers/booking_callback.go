package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// stepStartBooking handles the step of starting the booking process
func (h *BookingFormHandler) stepStartBooking(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string) error {
	if len(parts) < 2 {
		return h.sendError(query.Message.Chat.ID, "неверный формат запроса")
	}

	serviceID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	return h.startBooking(ctx, query, serviceID)
}

// stepMainMenu handles the step of going to the Main Menu
func (h *BookingFormHandler) stepMainMenu(ctx context.Context, userID int64, query *tgbotapi.CallbackQuery) error {
	if err := h.service.ClearSession(ctx, userID); err != nil {
		logger.Error("failed to clear session", zap.Error(err))
	}

	if err := h.sh.Handle(ctx, query); err != nil {
		logger.Error("failed to handle main_menu", zap.Error(err))
		return err
	}
	return nil
}

// stepDateSelect handles the date selection step
func (h *BookingFormHandler) stepDateSelect(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	state *botService.BookingState,
	parts []string,
) error {
	if len(parts) != 5 {
		return h.sendError(query.Message.Chat.ID, "неверный формат выбора даты")
	}

	if parts[1] != "select_date" {
		return h.sendError(query.Message.Chat.ID, "неверный формат")
	}

	startTime := strings.ReplaceAll(parts[3], ".", ":")
	endTime := strings.ReplaceAll(parts[4], ".", ":")

	slot := models.BoxAvailableSlot{
		Date:      parts[2],
		StartTime: startTime,
		EndTime:   endTime,
	}

	return h.dateSelection(ctx, query, state, slot)
}

// startBooking initiates the booking process
func (h *BookingFormHandler) startBooking(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	serviceID int64,
) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	msgID := query.Message.MessageID

	state, err := h.service.CreateSession(ctx, userID, serviceID)
	if err != nil {
		return err
	}

	logger.Info("Booking process started",
		zap.Int64("user_id", userID),
		zap.Int64("service_id", serviceID))

	return h.renderDateSelection(ctx, state, chatID, msgID)
}

// renderDateSelection displays the date selection step with buttons
func (h *BookingFormHandler) renderDateSelection(
	ctx context.Context,
	state *botService.BookingState,
	chatID int64,
	msgID int,
) error {
	slots, err := h.service.GetAvailableSlots(ctx, int64(state.ServiceID))
	if err != nil {
		logger.Error("failed to get dates from repository",
			zap.Error(err),
			zap.Int64("user_id", state.UserID))
		return h.sendError(chatID, "Ошибка получения слотов дат")
	}

	if len(slots) == 0 {
		msg := tgbotapi.NewEditMessageText(state.UserID, msgID,
			"На данный момент нет доступных слотов для бронирования",
		)
		keyboard := h.keyboard.FormNavigationKeyboard(botService.StepReturnInBoxList)
		msg.ReplyMarkup = &keyboard
		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
		return nil
	}

	keyboard := h.keyboard.DatesKeyboard(slots)
	messageText := "Выберите дату:\n"
	msg := tgbotapi.NewEditMessageText(state.UserID, msgID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	if _, err := h.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// dateSelection handles the user's date selection
func (h *BookingFormHandler) dateSelection(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	state *botService.BookingState,
	slot models.BoxAvailableSlot,
) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	msgID := query.Message.MessageID

	res, err := h.service.ProcessDateSelection(ctx, state, slot)
	if err != nil {
		logger.Error("date processing error", zap.Error(err))
		return h.sendError(chatID, "Не удалось обработать дату")
	}

	if !res {
		logger.Info("slot not available, restarting booking")
		return h.startBooking(ctx, query, state.ServiceID)
	}

	logger.Info("Date selected successfully",
		zap.Int64("user_id", userID),
		zap.String("date", state.SelectedSlot.Date),
		zap.String("start_time", state.SelectedSlot.StartTime),
		zap.String("end_time", state.SelectedSlot.EndTime))

	return h.renderNameInput(chatID, msgID)
}

// stepConfirmation processes the booking confirmation
func (h *BookingFormHandler) stepConfirmation(ctx context.Context, query *tgbotapi.CallbackQuery, state *botService.BookingState) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	bookingID, err := h.service.CreateBooking(ctx, state)
	if err != nil {
		logger.Error("booking saving error", zap.Error(err))
		return h.sendError(chatID, "Не удалось сохранить бронирование")
	}

	var messageText strings.Builder
	messageText.WriteString("Бронирование успешно создано!\n\n")
	messageText.WriteString("Номер бронирования: #")
	if _, err := fmt.Fprintf(&messageText, "%d", bookingID); err != nil {
		return fmt.Errorf("format booking id: %w", err)
	}
	messageText.WriteString("\nДата: ")
	messageText.WriteString(state.SelectedSlot.Date)
	messageText.WriteString("\nВремя: ")
	if _, err := fmt.Fprintf(&messageText, "%s - %s", state.SelectedSlot.StartTime, state.SelectedSlot.EndTime); err != nil {
		return fmt.Errorf("format booking id: %w", err)
	}

	messageText.WriteString("\nСтатус: Ожидает подтверждения\n\n")
	successMsg := messageText.String()

	msg := tgbotapi.NewMessage(chatID, successMsg)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = h.keyboard.MainMenuKeyboard()

	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	logger.Info("Booking confirmed",
		zap.Int64("booking_id", bookingID),
		zap.Int64("user_id", userID),
		zap.Int64("service_id", state.ServiceID))

	return nil
}
