package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/bot"
)

const cmdStart = "start"

type Handler struct {
	bot *bot.TelegramBot
}

func NewHandler(b *bot.TelegramBot) *Handler {
	return &Handler{
		bot: b,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	switch update.Message.Command() {
	case cmdStart:
		HandleStart(h.bot, update.Message)
	// в новые ветки добавлять вызовы функций обработчиков команд
	default:
		// todo
	}
}
