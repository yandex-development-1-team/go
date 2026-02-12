package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

const cmdStart = "start"

type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	// новые методы для bot api добавлять сюда, а реализовывать в go/internal/bot/bot.go
}

type MsgRateLimiter interface {
	Exec(ctx context.Context, chatID int64, f func() error) error
}

type Handler struct {
	bot   Bot
	msgRL MsgRateLimiter
}

func NewHandler(bot Bot, msgRL MsgRateLimiter) *Handler {
	return &Handler{
		bot:   bot,
		msgRL: msgRL,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	if msg := update.Message; msg != nil {
		if msg.IsCommand() {
			h.handleCommand(ctx, msg)
		}
		// todo
		return
	}
	if callbackQuery := update.CallbackQuery; callbackQuery != nil {
		// todo
	}
}

func (h *Handler) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	switch msg.Command() {
	case cmdStart:
		if err := h.msgRL.Exec(ctx, chatID, func() error { return HandleStart(h.bot, msg) }); err != nil {
			logger.Error("failed to handle /start", zap.Error(err))
		}
	// в новые ветки добавлять вызовы функций обработчиков команд
	default:
		// todo
	}
}
