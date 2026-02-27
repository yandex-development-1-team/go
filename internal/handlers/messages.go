package handlers

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/logger"
)

// MessageRouter router structure
type MessageRouter struct {
	bot         *tgbotapi.BotAPI
	sh          *StartHandler
	session     *repository.SessionRepository
	bookHandler *BookingFormHandler
}

// NewMessageRouter creates a new MessageRouter
func NewMessageRouter(
	bot *tgbotapi.BotAPI,
	sh *StartHandler,
	state *repository.SessionRepository,
	bookHandler *BookingFormHandler,
) *MessageRouter {
	return &MessageRouter{
		bot:         bot,
		sh:          sh,
		session:     state,
		bookHandler: bookHandler,
	}
}

// HandleMessage handles incoming messages
func (r *MessageRouter) HandleMessage(msg *tgbotapi.Message) {
	if msg.IsCommand() {
		r.handleCommand(msg)
		return
	}

	if msg.Text == "" {
		return
	}

	chatID := msg.Chat.ID
	userID := msg.From.ID

	ctxSession, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	state, err := r.session.GetSession(ctxSession, userID)
	if err != nil {
		logger.Error("Failed to get user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
	}

	ctxStep, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	switch state.CurrentState {
	case CallbackBookingPrefix:
		r.bookHandler.HandleTextMessage(ctxStep, userID, chatID, msg.Text)
	}
}

// handleCommand processes commands
func (r *MessageRouter) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		if err := r.sh.HandleStart(msg); err != nil {
			logger.Error("failed to handle /start", zap.Error(err))
		}

	default:
		// TODO
	}
}
