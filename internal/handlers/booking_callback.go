package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// stepStartBooking handles the step of starting the booking process
func (h *BookingFormHandler) stepStartBooking(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string) error {
	if len(parts) < 3 {
		return h.sendError(query.Message.Chat.ID, "неверный формат запроса")
	}
	visitType := parts[1]
	serviceID, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}
	return h.startBooking(ctx, query, serviceID, visitType)
}

func (h *BookingFormHandler) stepMainMenu(ctx context.Context, userID int64, msg *tgbotapi.Message) error {
	if err := h.service.ClearSession(ctx, userID); err != nil {
		logger.Error("failed to clear session", zap.Error(err))
	}

	if err := h.sh.HandleStart(ctx, msg); err != nil {
		logger.Error("failed to handle main_menu", zap.Error(err))
		return err
	}
	return nil
}

func (h *BookingFormHandler) stepDateSelect(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	state *botService.BookingState,
	parts []string,
) error {
	if len(parts) != 4 {
		return h.sendError(query.Message.Chat.ID, "неверный формат выбора даты")
	}

	return h.dateSelection(ctx, query, state, parts[3])
}

// startBooking initiates the booking process
func (h *BookingFormHandler) startBooking(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	serviceID int,
	visitType string,
) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	msgID := query.Message.MessageID

	state, err := h.service.CreateSession(ctx, userID, serviceID, visitType)
	if err != nil {
		return err
	}

	logger.Info("Booking process started",
		zap.Int64("user_id", userID),
		zap.Int("service_id", serviceID),
		zap.String("visit_type", visitType))

	return h.renderDateSelection(ctx, state, chatID, msgID)
}

// renderDateSelection displays the date selection step with buttons
func (h *BookingFormHandler) renderDateSelection(
	ctx context.Context,
	state *botService.BookingState,
	chatID int64,
	msgID int,
) error {
	availableDates, err := h.service.GetAvailableDates(ctx, state.ServiceID, state.VisitType)
	if err != nil {
		logger.Error("failed to get dates from repository",
			zap.Error(err),
			zap.Int64("user_id", state.UserID))
		return h.sendError(chatID, "Ошибка получения дат")
	}

	if len(availableDates) == 0 {
		msg := tgbotapi.NewEditMessageText(state.UserID, msgID,
			"На данный момент нет доступных дат для бронирования",
		)
		msg.ReplyMarkup = h.keyboard.MainMenuKeyboard()
		_, err := h.bot.Send(msg)
		return err
	}

	dates := make([]time.Time, 0, len(availableDates))
	for _, dateStr := range availableDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dates = append(dates, date)
		}
	}

	keyboard := h.keyboard.DatesKeyboard(state.VisitType, dates)
	messageText := "Выберите дату:\n"
	msg := tgbotapi.NewEditMessageText(state.UserID, msgID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	_, err = h.bot.Send(msg)
	return err
}

// DateSelection handles the user's date selection
func (h *BookingFormHandler) dateSelection(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
	state *botService.BookingState,
	dateStr string,
) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	msgID := query.Message.MessageID

	res, err := h.service.ProcessDateSelection(ctx, state, dateStr)
	if err != nil {
		logger.Error("date processing error", zap.Error(err))
		return h.sendError(chatID, "Не удалось обработать дату")
	}

	if !res {
		return h.startBooking(ctx, query, state.ServiceID, state.VisitType)
	}

	logger.Info("Date selected",
		zap.Int64("user_id", userID),
		zap.Time("date", state.SelectedDate))

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
	messageText.WriteString(state.SelectedDate.Format("02.01.2006"))
	messageText.WriteString("\nСтатус: Ожидает подтверждения\n\n")
	successMsg := messageText.String()

	msg := tgbotapi.NewMessage(chatID, successMsg)
	msg.ParseMode = "Markdown"

	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	logger.Info("Booking confirmed",
		zap.Int64("booking_id", bookingID),
		zap.Int64("user_id", userID),
		zap.Int("service_id", state.ServiceID))

	backMsg := tgbotapi.NewMessage(chatID, "Вернуться в меню:")
	backMsg.ReplyMarkup = h.keyboard.MainMenuKeyboard()
	_, err = h.bot.Send(backMsg)

	return err
}
