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

// BotAPI interface for working with Telegram API
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// BookingFormHandler processes the booking form
type BookingFormHandler struct {
	bot      BotAPI
	sh       *StartHandler
	bs       *BoxSolutionsHandler
	service  *botService.BookingService
	keyboard *KeyboardService
}

// NewBookingFormHandler creates a new instance of the booking form handler
func NewBookingFormHandler(
	bot *tgbotapi.BotAPI,
	service *botService.BookingService,
	sh *StartHandler,
	bs *BoxSolutionsHandler,
	keyboard *KeyboardService,
) *BookingFormHandler {

	return &BookingFormHandler{
		bot:      bot,
		service:  service,
		sh:       sh,
		bs:       bs,
		keyboard: keyboard,
	}
}

// Handle handles callback requests from inline booking buttons
func (h *BookingFormHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) (err error) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	if _, err := h.bot.Request(tgbotapi.NewCallback(query.ID, "")); err != nil {
		logger.Error("answer callback query", zap.Error(err), zap.String("callback_id", query.ID))
		return fmt.Errorf("answer callback query: %w", err)
	}

	logger.Info("Booking callback received",
		zap.Int64("user_id", userID),
		zap.String("data", query.Data))

	parts := strings.Split(query.Data, ":")
	state := h.service.GetBookingState(ctx, userID)
	action := h.GetAction(state, parts)

	if state != nil && state.OldMessageID != nil {
		delTgMessage(h.bot, &tgbotapi.Message{MessageID: *state.OldMessageID, Chat: &tgbotapi.Chat{ID: chatID}})
	} else if action == botService.StepStartBooking {
		delTgMessage(h.bot, query.Message)
	}

	if parts[1] == "back" {
		return h.handleBack(ctx, state, query, parts)
	}

	switch action {
	case botService.StepStartBooking:
		return h.stepStartBooking(ctx, query, parts)

	case botService.StepSelectDate:
		if state == nil {
			return h.sendError(chatID, "сессия не найдена")
		}
		return h.stepDateSelect(ctx, query, state, parts)

	case botService.StepConfirmation:
		if state == nil {
			return h.sendError(chatID, "сессия не найдена")
		}
		return h.stepConfirmation(ctx, query, state)

	case botService.StepMainMenu:
		return h.stepMainMenu(ctx, userID, query)

	default:
		logger.Warn("unknown action", zap.Int("Action", action))
		return h.sendError(chatID, "неизвестное действие")
	}
}

// HandleTextMessage processes text messages to fill out the booking form
func (h *BookingFormHandler) HandleTextMessage(ctx context.Context, msg *tgbotapi.Message) error {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	state := h.service.GetBookingState(ctx, userID)
	if state == nil {
		return h.sendError(chatID, "Ошибка при получении статуса")
	}

	if state.OldMessageID != nil {
		delTgMessage(h.bot, &tgbotapi.Message{MessageID: *state.OldMessageID, Chat: &tgbotapi.Chat{ID: chatID}})
	}

	logger.Debug("Text input processing",
		zap.Int64("user_id", userID),
		zap.Int("step", int(state.Step)),
		zap.String("text", text))

	switch state.Step {
	case botService.StepEnterName:
		return h.stepNameInput(ctx, state, msg)

	case botService.StepEnterOrg:
		return h.stepOrganizationInput(ctx, state, msg)

	case botService.StepEnterPosition:
		return h.stepPositionInput(ctx, state, msg)

	default:
		logger.Warn("unknown action", zap.Int("Action", state.Step))
		return h.sendError(chatID, "неизвестное действие")
	}
}

// GetAction returns the status of the booking process
func (h *BookingFormHandler) GetAction(state *botService.BookingState, parts []string) int {
	if len(parts) < 2 {
		logger.Error("invalid parameters")
		return botService.StepMainMenu
	}

	if parts[1] == "main_menu" {
		return botService.StepMainMenu
	}

	if state == nil {
		return botService.StepStartBooking
	}
	return state.Step
}

// handleBack returns the state to the previous step
func (h *BookingFormHandler) handleBack(
	ctx context.Context,
	state *botService.BookingState,
	query *tgbotapi.CallbackQuery,
	parts []string,
) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	if len(parts) == 4 {
		parts[2] = strconv.Itoa(botService.StepReturnInBoxList)
	}

	if len(parts) < 3 {
		return h.sendError(chatID, "Неверный формат кнопки Назад")
	}

	if state == nil || state.OldMessageID == nil {
		return h.sendError(chatID, "Сессии пользователя не существует")
	}

	targetStep, err := strconv.Atoi(parts[2])
	if err != nil {
		return h.sendError(chatID, "Неверный шаг")
	}

	if query.Message.MessageID != *state.OldMessageID {
		return h.sendError(chatID, "Ошибка в сессии")
	}

	switch targetStep {
	case botService.StepReturnInBoxList:
		if err := h.service.ClearSession(ctx, userID); err != nil {
			return fmt.Errorf("clear session: %w", err)
		}

		query.Data = fmt.Sprintf("%s:page:%s", CallbackBoxSolutions, parts[3])
		return h.bs.Handle(ctx, query)

	case botService.StepStartBooking:
		state.Step = botService.StepSelectDate
		state.SelectedSlot = models.BoxAvailableSlot{}
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderDateSelection(ctx, state, chatID, "1")

	case botService.StepEnterName:
		state.Step = botService.StepEnterName
		query.Message.Text = state.GuestName
		state.GuestName = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderNameInput(ctx, query, state)

	case botService.StepEnterOrg:
		state.Step = botService.StepEnterOrg
		query.Message.Text = state.GuestOrganization
		state.GuestOrganization = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderOrganizationInput(ctx, chatID, state)

	case botService.StepEnterPosition:
		state.Step = botService.StepEnterPosition
		query.Message.Text = state.GuestPosition
		state.GuestPosition = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderPositionInput(ctx, chatID, state)

	default:
		return h.sendError(chatID, "Нельзя вернуться на этот шаг")
	}
}

// sendError sends an error message
func (h *BookingFormHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %s", errorMsg))
	if _, err := h.bot.Send(msg); err != nil {
		return err
	}
	return nil
}
