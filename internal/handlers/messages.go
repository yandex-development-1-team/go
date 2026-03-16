package handlers

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/logger"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// MessageRouter router structure
type MessageRouter struct {
	bot         *tgbotapi.BotAPI
	sh          *StartHandler
	session     repository.SessionRepository
	bookHandler *BookingFormHandler
	msgRL       MsgRateLimiter
}

// NewMessageRouter creates a new MessageRouter
func NewMessageRouter(
	bot *tgbotapi.BotAPI,
	sh *StartHandler,
	session repository.SessionRepository,
	bookHandler *BookingFormHandler,
	msgRL MsgRateLimiter,
) *MessageRouter {
	return &MessageRouter{
		bot:         bot,
		sh:          sh,
		session:     session,
		bookHandler: bookHandler,
		msgRL:       msgRL,
	}
}

// HandleMessage handles incoming messages
func (r *MessageRouter) HandleMessage(ctx context.Context, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		r.handleCommand(ctx, msg)
		return
	}

	if msg.Text == "" {
		return
	}

	userID := msg.From.ID
	// chatID := msg.Chat.ID
	// text := msg.Text

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
	case botService.CallbackBookingPrefix:
		r.bookHandler.HandleTextMessage(ctxStep, msg)
	}
}

// handleCommand processes commands
func (r *MessageRouter) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		if err := r.msgRL.Exec(ctx, msg.Chat.ID, func() error { return r.sh.HandleStart(ctx, msg) }); err != nil {
			logger.Error("failed to handle /start", zap.Error(err))
		}
	}
}
