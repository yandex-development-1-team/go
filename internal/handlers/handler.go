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

type RateLimiter interface {
	Exec(ctx context.Context, f func() error) error
}

type Handler struct {
	bot          Bot
	msgRL, apiRL RateLimiter
}

func NewHandler(bot Bot, msgRL, apiRL RateLimiter) *Handler {
	return &Handler{
		bot:   bot,
		msgRL: msgRL,
		apiRL: apiRL,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	if msg := update.Message; msg != nil {
		if msg.IsCommand() {
			switch msg.Command() {
			case cmdStart:
				if err := h.msgRL.Exec(ctx, func() error { return HandleStart(h.bot, msg) }); err != nil {
					logger.Error("failed to handle /start", zap.Error(err))
				}
			// в новые ветки добавлять вызовы функций обработчиков команд
			default:
				// todo
			}
		}
		// todo
	}
	if callbackQuery := update.CallbackQuery; callbackQuery != nil {
		// todo
	}
}
