package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/yandex-development-1-team/go/internal/database/inmemory"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

// getAction returns the status of the booking process
func (h *BookingFormHandler) getAction(userID int64, parts []string) int {
	if len(parts) < 2 {
		return StepStartBooking
	}

	if parts[1] == "main_menu" {
		return StepMainMenu
	}

	state, err := h.state.Get(userID)
	if err != nil {
		return StepStartBooking
	}
	return state.Step
}

// stepStartBooking handles the step of starting the booking process
func (h *BookingFormHandler) stepStartBooking(
	ctx context.Context,
	userID int64,
	chatID int64,
	msgID int,
	parts []string,
) error {
	if len(parts) < 3 {
		return h.sendError(chatID, "неверный формат запроса")
	}
	visitType := parts[1]
	serviceID, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}
	return h.startBooking(ctx, userID, chatID, serviceID, visitType, msgID)
}

func (h *BookingFormHandler) stepMainMenu(msg *tgbotapi.Message) error {
	if err := h.sh.HandleStart(msg); err != nil {
		logger.Error("failed to handle main_menu", zap.Error(err))
		return err
	}

	return nil
}

func (h *BookingFormHandler) stepDateSelect(
	ctx context.Context,
	userID int64,
	chatID int64,
	msgID int,
	parts []string,
) error {
	if len(parts) != 4 {
		return h.sendError(chatID, "неверный формат выбора даты")
	}

	return h.handleDateSelection(ctx, userID, chatID, msgID, parts[2], parts[3])
}

// startBooking initiates the booking process
func (h *BookingFormHandler) startBooking(
	ctx context.Context,
	userID int64,
	chatID int64,
	serviceID int,
	visitType string,
	msgID int,
) error {
	// Create state
	state := &inmemory.BookingState{
		UserID:    userID,
		ServiceID: serviceID,
		VisitType: visitType,
		Step:      StepSelectDate,
		CreatedAt: time.Now(),
	}

	// Save state
	if err := h.state.Save(userID, state); err != nil {
		return errors.Wrap(err, "failed to save state")
	}

	logger.Info("Booking process started",
		zap.Int64("user_id", userID),
		zap.Int("service_id", serviceID),
		zap.String("visit_type", visitType))

	// Showing the available dates
	return h.renderDateSelection(ctx, state, chatID, msgID)
}

// renderDateSelection displays the date selection step with buttons
func (h *BookingFormHandler) renderDateSelection(
	ctx context.Context,
	state *inmemory.BookingState,
	chatID int64,
	msgID int,
) error {

	// Getting dates from the array (temporary solution)
	availableDates, err := GetAvailableDates(ctx, state.ServiceID, state.VisitType, 30)
	if err != nil {
		logger.Error("failed to get dates from cache",
			zap.Error(err),
			zap.Int64("user_id", state.UserID))
		return h.sendError(chatID, "Ошибка получения дат")
	}

	if len(availableDates) == 0 {
		msg := tgbotapi.NewEditMessageText(state.UserID, msgID,
			"На данный момент нет доступных дат для бронирования",
		)
		msg.ReplyMarkup = h.keyboard.createMainMenuKeyboard()
		_, err := h.bot.Send(msg)
		return err
	}

	// Creating a keyboard with dates
	keyboard := h.keyboard.createDatesKeyboard(availableDates)
	messageText := "Выберите дату:\n" //fmt.Sprintf("Выберите дату:\n")
	msg := tgbotapi.NewEditMessageText(state.UserID, msgID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	_, err = h.bot.Send(msg)
	return err
}

// handleDateSelection handles the user's date selection
func (h *BookingFormHandler) handleDateSelection(
	ctx context.Context,
	userID int64,
	chatID int64,
	msgID int,
	visitType string,
	dateStr string,
) error {
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		logger.Error("date parsing error", zap.Error(err), zap.String("date_str", dateStr))
		return h.sendError(chatID, "Неверный формат даты")
	}

	state, err := h.state.Get(userID)
	if err != nil {
		return h.sendError(chatID, "Сессия бронирования устарела")
	}

	available, err := IsDateAvailable(ctx, state.ServiceID, selectedDate, visitType)
	if err != nil {
		logger.Error("date availability verification error", zap.Error(err))
		return h.sendError(chatID, "Не удалось проверить доступность даты")
	}

	if !available {
		state.Step = StepStartBooking
		if err := h.state.Save(userID, state); err != nil {
			logger.Error("status save error", zap.Error(err))
		}
		return h.startBooking(ctx, userID, chatID, state.ServiceID, visitType, msgID)
	}

	state.SelectedDate = selectedDate
	state.Step = StepEnterName

	if err := h.state.Save(userID, state); err != nil {
		logger.Error("status save error", zap.Error(err))
	}

	logger.Info("Date selected",
		zap.Int64("user_id", userID),
		zap.Time("date", selectedDate))

	return h.renderNameInput(chatID, msgID)
}

// renderNameInput displays the step of entering the full name
func (h *BookingFormHandler) renderNameInput(
	chatID int64,
	msgID int,
) error {
	var messageText strings.Builder
	messageText.WriteString("*Введите ФИО*\n\n")
	messageText.WriteString("Формат: Фамилия Имя Отчество\n")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", "book:main_menu"),
		),
	)

	msg := tgbotapi.NewEditMessageText(chatID, msgID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	_, err := h.bot.Send(msg)
	return err
}

// handleConfirmation processes the booking confirmation
func (h *BookingFormHandler) stepConfirmation(
	ctx context.Context,
	userID int64,
	chatID int64,
) error {
	state, err := h.state.Get(userID)
	if err != nil || state == nil {
		return h.sendError(chatID, "Сессия бронирования устарела. Начните заново.")
	}

	booking := &models.Booking{
		UserID:            state.UserID,
		ServiceID:         int16(state.ServiceID),
		BookingDate:       state.SelectedDate,
		BookingTime:       nil,
		GuestName:         state.GuestName,
		GuestOrganization: state.GuestOrganization,
		GuestPosition:     state.GuestPosition,
		VisitType:         state.VisitType,
		Status:            "confirmation",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	bookingID, err := h.repo.CreateBooking(ctx, booking)
	if err != nil {
		logger.Error("booking saving error", zap.Error(err))
		return h.sendError(chatID, "Не удалось сохранить бронирование")
	}

	h.state.Delete(userID)

	var messageText strings.Builder
	messageText.WriteString("Бронирование успешно создано!\n\n")
	messageText.WriteString("Номер бронирования: #")
	messageText.WriteString(fmt.Sprintf("%d", bookingID))
	messageText.WriteString("\nДата: ")
	messageText.WriteString(state.SelectedDate.Format("02.01.2006"))
	messageText.WriteString("\nСтатус: Ожидает подтверждения\n\n")
	successMsg := messageText.String()

	msg := tgbotapi.NewMessage(chatID, successMsg)
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)

	logger.Info("Booking confirmed",
		zap.Int64("booking_id", bookingID),
		zap.Int64("user_id", userID),
		zap.Int("service_id", state.ServiceID))

	backMsg := tgbotapi.NewMessage(chatID, "Вернуться в меню:")
	backMsg.ReplyMarkup = h.keyboard.createMainMenuKeyboard()
	_, err = h.bot.Send(backMsg)

	return err
}
