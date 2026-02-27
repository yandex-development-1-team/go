package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/database/inmemory"
	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/logger"
)

// State constants of the booking process
const (
	StepStartBooking = iota
	StepSelectDate
	StepEnterName
	StepEnterOrg
	StepEnterPosition
	StepConfirmation
	StepMainMenu
)

// CallbackBookingPrefix prefix for callback data
const CallbackBookingPrefix = "book"

// BotAPI интерфейс для работы с Telegram API
// Нужен для создания мока бота в тестах
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// BookingFormHandler processes the booking form
type BookingFormHandler struct {
	//bot *tgbotapi.BotAPI
	bot      BotAPI
	sh       *StartHandler
	state    *inmemory.BookingStateStorage
	session  *repository.SessionRepository
	repo     *repository.BookingRepo
	keyboard *KeyboardService
}

// NewBookingFormHandler creates a new instance of the booking form handler
func NewBookingFormHandler(
	bot *tgbotapi.BotAPI,
	state *inmemory.BookingStateStorage,
	session *repository.SessionRepository,
	repo *repository.BookingRepo,
	keyboard *KeyboardService,
) *BookingFormHandler {
	return &BookingFormHandler{
		bot:      bot,
		state:    state,
		session:  session,
		repo:     repo,
		keyboard: keyboard,
	}
}

// Handle handles callback requests from inline booking buttons
func (h *BookingFormHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	msgID := query.Message.MessageID
	callbackData := query.Data

	logger.Info("Booking callback received",
		zap.Int64("user_id", userID),
		zap.String("data", callbackData))

	// Sending a response to remove the "clock" from the button
	h.bot.Request(tgbotapi.NewCallback(query.ID, ""))

	parts := strings.Split(callbackData, ":")
	action := h.getAction(userID, parts)
	h.saveSession(ctx, userID, CallbackBookingPrefix, parts)

	// Routing by type of action
	switch action {
	case StepStartBooking:
		return h.stepStartBooking(ctx, userID, chatID, msgID, parts)

	case StepSelectDate:
		// Format: book:date:{visitType}:{YYYY-MM-DD}
		return h.stepDateSelect(ctx, userID, chatID, msgID, parts)

	case StepConfirmation:
		// Format: book:confirm
		return h.stepConfirmation(ctx, userID, chatID)

	case StepMainMenu:
		// Format: book:main_menu
		return h.stepMainMenu(query.Message)

	default:
		logger.Warn("unknown action", zap.Int("Action", action))
		return h.sendError(chatID, "неизвестное действие")
	}
}

// HandleTextMessage processes text messages to fill out the booking form
func (h *BookingFormHandler) HandleTextMessage(
	ctx context.Context,
	userID int64,
	chatID int64,
	text string,
) error {
	state, err := h.state.Get(userID)
	if err != nil {
		logger.Error("error receiving the status", zap.Error(err))
		return nil
	}

	logger.Debug("Text input processing",
		zap.Int64("user_id", userID),
		zap.Int("step", int(state.Step)),
		zap.String("text", text))

	h.saveSession(ctx, userID, CallbackBookingPrefix, text)

	// Process it depending on the current step.
	switch state.Step {
	case StepEnterName:
		return h.stepNameInput(ctx, state, chatID, text)

	case StepEnterOrg:
		return h.stepOrganizationInput(ctx, state, chatID, text)

	case StepEnterPosition:
		return h.stepPositionInput(state, chatID, text)

	default:
		logger.Warn("unknown action", zap.Int("Action", state.Step))
		return h.sendError(chatID, "неизвестное действие")
	}
}

// sendError sends an error message
func (h *BookingFormHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %s", errorMsg))
	_, err := h.bot.Send(msg)
	return err
}

func (h *BookingFormHandler) saveSession(
	ctx context.Context,
	userID int64,
	state string,
	stateData interface{},
) {
	data := map[string]interface{}{
		"data": stateData,
	}

	err := h.session.SaveSession(ctx, userID, state, data)
	if err != nil {
		logger.Error("failed to save user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
	}
}
