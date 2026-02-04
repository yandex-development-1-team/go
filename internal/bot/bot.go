package bot

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type TelegramBot struct {
	Api    *tgbotapi.BotAPI
	Logger *zap.Logger
}

func NewTelegramBot(token string, logger *zap.Logger) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	user, err := bot.GetMe()
	if err != nil {
		return nil, err
	}
	logger.Info("telegram bot authorized on account", zap.String("bot_name", user.UserName), zap.Int64("ID", user.ID))

	bot.Debug = true

	logger.Info("telegram bot has been initialized")
	return &TelegramBot{
		Api:    bot,
		Logger: logger,
	}, nil
}

func (b *TelegramBot) GetUpdates(timeout time.Duration) tgbotapi.UpdatesChannel {
	updates := b.Api.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout:        int(timeout.Seconds()),
		AllowedUpdates: []string{"message", "callback_query", "my_chat_member"},
	})
	return updates
}

func (b *TelegramBot) Shutdown(ctx context.Context) {
	b.Api.StopReceivingUpdates()
}
