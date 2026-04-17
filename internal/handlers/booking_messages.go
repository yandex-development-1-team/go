package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// renderNameInput displays the step of entering the full name
func (h *BookingFormHandler) renderNameInput(ctx context.Context, query *tgbotapi.CallbackQuery, state *botService.BookingState) error {
	chatID := query.Message.Chat.ID

	var messageText strings.Builder
	messageText.WriteString("*Введите ФИО*\n\n")
	messageText.WriteString("Формат: Фамилия Имя Отчество\n")

	keyboard := h.keyboard.FormNavigationKeyboard(botService.StepStartBooking)
	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	sent, err := h.bot.Send(msg)
	if err != nil {
		return err
	}

	state.OldMessageID = &sent.MessageID
	err = h.service.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// stepNameInput processes the input of the full name
func (h *BookingFormHandler) stepNameInput(
	ctx context.Context,
	state *botService.BookingState,
	msg *tgbotapi.Message,
) error {
	chatID := msg.Chat.ID
	text := msg.Text

	if err := h.service.ValidateAndSetName(ctx, state, text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации ФИО*\n\n%s\n\nВведите ФИО еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard(botService.StepSelectDate)
		sent, err := h.bot.Send(msg)
		if err != nil {
			return err
		}

		state.OldMessageID = &sent.MessageID
		err = h.service.SaveSession(ctx, state.UserID, *state)
		if err != nil {
			return err
		}
		return nil
	}

	logger.Info("Full name entered",
		zap.Int64("user_id", state.UserID),
		zap.String("name", state.GuestName))

	return h.renderOrganizationInput(ctx, chatID, state)
}

// renderOrganizationInput displays the step of entering the organization
func (h *BookingFormHandler) renderOrganizationInput(ctx context.Context, chatID int64, state *botService.BookingState) error {
	var messageText strings.Builder
	messageText.WriteString("Введите организацию\n\n")
	messageText.WriteString("Можно использовать буквы, цифры, кавычки\n\n")
	messageText.WriteString("Пример: ООО \"Ромашка\"")

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard(botService.StepEnterName)

	sent, err := h.bot.Send(msg)
	if err != nil {
		return err
	}

	state.OldMessageID = &sent.MessageID
	err = h.service.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// stepOrganizationInput processes the organization's input
func (h *BookingFormHandler) stepOrganizationInput(
	ctx context.Context,
	state *botService.BookingState,
	msg *tgbotapi.Message,
) error {
	chatID := msg.Chat.ID
	text := msg.Text

	if err := h.service.ValidateAndSetOrganization(ctx, state, text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации организации*\n\n%s\n\nВведите название организации еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard(botService.StepEnterName)

		sent, err := h.bot.Send(msg)
		if err != nil {
			return err
		}

		state.OldMessageID = &sent.MessageID
		err = h.service.SaveSession(ctx, state.UserID, *state)
		if err != nil {
			return err
		}
		return nil
	}

	logger.Info("The organization has been introduced",
		zap.Int64("user_id", state.UserID),
		zap.String("org", state.GuestOrganization))

	return h.renderPositionInput(ctx, chatID, state)
}

// renderPositionInput displays the step of entering the position
func (h *BookingFormHandler) renderPositionInput(ctx context.Context, chatID int64, state *botService.BookingState) error {
	var messageText strings.Builder
	messageText.WriteString("Введите должность\n\n")
	messageText.WriteString("Пример: Менеджер по продажам")

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard(botService.StepEnterOrg)

	sent, err := h.bot.Send(msg)
	if err != nil {
		return err
	}

	state.OldMessageID = &sent.MessageID
	err = h.service.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// stepPositionInput processes the entry of a position
func (h *BookingFormHandler) stepPositionInput(
	ctx context.Context,
	state *botService.BookingState,
	msg *tgbotapi.Message,
) error {
	chatID := msg.Chat.ID
	text := msg.Text

	if err := h.service.ValidateAndSetPosition(ctx, state, text); err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"*Ошибка валидации должности*\n\n%s\n\nВведите должность еще раз:",
			err.Error(),
		))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = h.keyboard.FormNavigationKeyboard(botService.StepEnterOrg)

		sent, err := h.bot.Send(msg)
		if err != nil {
			return err
		}

		state.OldMessageID = &sent.MessageID
		err = h.service.SaveSession(ctx, state.UserID, *state)
		if err != nil {
			return err
		}
		return nil
	}

	logger.Info("The position has been introduced",
		zap.Int64("user_id", state.UserID),
		zap.String("position", state.GuestPosition))

	return h.renderConfirmation(ctx, chatID, state)
}

// renderConfirmation displays the booking confirmation step
func (h *BookingFormHandler) renderConfirmation(ctx context.Context, chatID int64, state *botService.BookingState) error {
	var messageText strings.Builder
	messageText.WriteString("Подтверждение бронирования\n\n")
	if _, err := fmt.Fprintf(&messageText, "Название: %s\n", state.ServiceName); err != nil {
		return fmt.Errorf("format confirmation title: %w", err)
	}
	if _, err := fmt.Fprintf(&messageText, "Дата: %s\n", state.SelectedSlot.Date); err != nil {
		return fmt.Errorf("format confirmation date: %w", err)
	}
	if _, err := fmt.Fprintf(&messageText, "Время: %s - %s\n", state.SelectedSlot.StartTime, state.SelectedSlot.EndTime); err != nil {
		return fmt.Errorf("format confirmation date: %w", err)
	}
	if _, err := fmt.Fprintf(&messageText, "ФИО: %s\n", state.GuestName); err != nil {
		return fmt.Errorf("format confirmation name: %w", err)
	}
	if _, err := fmt.Fprintf(&messageText, "Организация: %s\n", state.GuestOrganization); err != nil {
		return fmt.Errorf("format confirmation org: %w", err)
	}
	if _, err := fmt.Fprintf(&messageText, "Должность: %s\n\n", state.GuestPosition); err != nil {
		return fmt.Errorf("format confirmation position: %w", err)
	}
	messageText.WriteString("Проверьте правильность введенных данных\n")

	keyboard := h.keyboard.ConfirmationKeyboard(botService.StepEnterPosition)

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	sent, err := h.bot.Send(msg)
	if err != nil {
		return err
	}

	state.OldMessageID = &sent.MessageID
	err = h.service.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}
