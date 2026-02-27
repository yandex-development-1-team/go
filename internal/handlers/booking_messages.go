package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/database/inmemory"
	"github.com/yandex-development-1-team/go/internal/handlers/validation"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

// processNameInput processes the input of the full name
func (h *BookingFormHandler) stepNameInput(
	ctx context.Context,
	state *inmemory.BookingState,
	chatID int64,
	text string,
) error {
	if err := validation.Name(text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации ФИО*\n\n%s\n\nВведите ФИО еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard()
		_, sendErr := h.bot.Send(msg)
		return sendErr
	}

	state.GuestName = strings.TrimSpace(text)
	state.Step = StepEnterOrg

	if err := h.state.Save(state.UserID, state); err != nil {
		logger.Error("status update error", zap.Error(err))
	}

	logger.Info("Full name entered",
		zap.Int64("user_id", state.UserID),
		zap.String("name", state.GuestName))

	return h.renderOrganizationInput(chatID)
}

// renderOrganizationInput displays the step of entering the organization
func (h *BookingFormHandler) renderOrganizationInput(
	chatID int64,
) error {
	var messageText strings.Builder
	messageText.WriteString("Введите организацию\n\n")
	messageText.WriteString("Можно использовать буквы, цифры, кавычки\n\n")
	messageText.WriteString("Пример: ООО \"Ромашка\"")

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard()

	_, err := h.bot.Send(msg)
	return err
}

// processOrganizationInput processes the organization's input
func (h *BookingFormHandler) stepOrganizationInput(
	ctx context.Context,
	state *inmemory.BookingState,
	chatID int64,
	text string,
) error {
	if err := validation.Organization(text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации организации*\n\n%s\n\nВведите название организации еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard()
		_, sendErr := h.bot.Send(msg)
		return sendErr
	}

	state.GuestOrganization = strings.TrimSpace(text)
	state.Step = StepEnterPosition

	if err := h.state.Save(state.UserID, state); err != nil {
		logger.Error("status update error", zap.Error(err))
	}

	logger.Info("The organization has been introduced",
		zap.Int64("user_id", state.UserID),
		zap.String("org", state.GuestOrganization))

	return h.renderPositionInput(chatID)
}

// renderPositionInput displays the step of entering the position
func (h *BookingFormHandler) renderPositionInput(
	chatID int64,
) error {
	var messageText strings.Builder
	messageText.WriteString("Введите должность\n\n")
	messageText.WriteString("Пример: Менеджер по продажам")

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard()

	_, err := h.bot.Send(msg)
	return err
}

// processPositionInput processes the entry of a position
func (h *BookingFormHandler) stepPositionInput(
	state *inmemory.BookingState,
	chatID int64,
	text string,
) error {
	if err := validation.Position(text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации должности*\n\n%s\n\nВведите должность еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard()
		_, sendErr := h.bot.Send(msg)
		return sendErr
	}

	state.GuestPosition = strings.TrimSpace(text)
	state.Step = StepConfirmation

	if err := h.state.Save(state.UserID, state); err != nil {
		logger.Error("status update error", zap.Error(err))
	}

	logger.Info("The position has been introduced",
		zap.Int64("user_id", state.UserID),
		zap.String("position", state.GuestPosition))

	return h.renderConfirmation(chatID, state)
}

// renderConfirmation displays the booking confirmation step
func (h *BookingFormHandler) renderConfirmation(
	chatID int64,
	state *inmemory.BookingState,
) error {
	var messageText strings.Builder
	messageText.WriteString("Подтверждение бронирования\n\n")
	messageText.WriteString(fmt.Sprintf("Дата: %s\n", state.SelectedDate.Format("02.01.2006")))
	messageText.WriteString(fmt.Sprintf("ФИО: %s\n", state.GuestName))
	messageText.WriteString(fmt.Sprintf("Организация: %s\n", state.GuestOrganization))
	messageText.WriteString(fmt.Sprintf("Должность: %s\n\n", state.GuestPosition))
	messageText.WriteString("Проверьте правильность введенных данных\n")

	keyboard := h.keyboard.ConfirmationKeyboard()

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}
