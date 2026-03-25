package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
)

type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type MsgRateLimiter interface {
	Exec(ctx context.Context, chatID int64, f func() error) error
}

type Handler struct {
	bot            Bot
	msgRL          MsgRateLimiter
	msgRouter      *MessageRouter
	callbackRouter *CallbackRouter
}

func NewHandler(bot Bot, msgRL MsgRateLimiter, msgRouter *MessageRouter, callbackRouter *CallbackRouter) *Handler {
	return &Handler{
		bot:            bot,
		msgRL:          msgRL,
		msgRouter:      msgRouter,
		callbackRouter: callbackRouter,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	activeUsers := getActiveUsersCount(ctx)
	metrics.SetActiveUsers(activeUsers)

	if msg := update.Message; msg != nil {
		h.msgRouter.HandleMessage(ctx, msg)
	}
	if callbackQuery := update.CallbackQuery; callbackQuery != nil {
		if err := HandleCallback(h.callbackRouter, update.CallbackQuery); err != nil {
			logger.Error("callback handling", zap.Error(err))
		}
	}
}

func getActiveUsersCount(_ context.Context) int {
	return 1
}
