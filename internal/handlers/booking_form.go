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

// BotAPI interface for working with Telegram API
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// BookingFormHandler processes the booking form
type BookingFormHandler struct {
	bot      BotAPI
	sh       *StartHandler
	service  *botService.BookingService
	keyboard *KeyboardService
}

// NewBookingFormHandler creates a new instance of the booking form handler
func NewBookingFormHandler(
	bot *tgbotapi.BotAPI,
	service *botService.BookingService,
	sh *StartHandler,
	keyboard *KeyboardService,
) *BookingFormHandler {

	return &BookingFormHandler{
		bot:      bot,
		service:  service,
		sh:       sh,
		keyboard: keyboard,
	}
}

// Handle handles callback requests from inline booking buttons
func (h *BookingFormHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) (err error) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	h.bot.Request(tgbotapi.NewCallback(query.ID, ""))

	logger.Info("Booking callback received",
		zap.Int64("user_id", userID),
		zap.String("data", query.Data))

	parts := strings.Split(query.Data, ":")
	state := h.service.GetBookingState(ctx, userID)
	action := h.GetAction(state, parts)

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
		return h.stepMainMenu(ctx, userID, query.Message)

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

	logger.Debug("Text input processing",
		zap.Int64("user_id", userID),
		zap.Int("step", int(state.Step)),
		zap.String("text", text))

	switch state.Step {
	case botService.StepEnterName:
		return h.stepNameInput(ctx, state, chatID, text)

	case botService.StepEnterOrg:
		return h.stepOrganizationInput(ctx, state, chatID, text)

	case botService.StepEnterPosition:
		return h.stepPositionInput(ctx, state, chatID, text)

	default:
		logger.Warn("unknown action", zap.Int("Action", state.Step))
		return h.sendError(chatID, "неизвестное действие")
	}
}

// GetAction returns the status of the booking process
func (h *BookingFormHandler) GetAction(
	//ctx context.Context,
	state *botService.BookingState,
	//query *tgbotapi.CallbackQuery,
	parts []string,
) int {
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
	msgID := query.Message.MessageID

	if len(parts) < 3 {
		return h.sendError(chatID, "Неверный формат кнопки Назад")
	}

	targetStep, err := strconv.Atoi(parts[2])
	if err != nil {
		return h.sendError(chatID, "Неверный шаг")
	}

	switch targetStep {
	case botService.StepStartBooking:
		state.Step = botService.StepSelectDate
		state.SelectedDate = time.Time{}
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderDateSelection(ctx, state, chatID, msgID)

	case botService.StepEnterName:
		state.Step = botService.StepEnterName
		query.Message.Text = state.GuestName
		state.GuestName = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderNameInput(chatID, msgID)

	case botService.StepEnterOrg:
		state.Step = botService.StepEnterOrg
		query.Message.Text = state.GuestOrganization
		state.GuestOrganization = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderOrganizationInput(chatID)

	case botService.StepEnterPosition:
		state.Step = botService.StepEnterPosition
		query.Message.Text = state.GuestPosition
		state.GuestPosition = ""
		if err := h.service.SaveSession(ctx, userID, *state); err != nil {
			return err
		}
		return h.renderPositionInput(chatID)

	default:
		return h.sendError(chatID, "Нельзя вернуться на этот шаг")
	}
}

// sendError sends an error message
func (h *BookingFormHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %s", errorMsg))
	_, err := h.bot.Send(msg)
	return err
}
