package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

type TelegramBot struct {
	Api *tgbotapi.BotAPI
}

func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	user, err := bot.GetMe()
	if err != nil {
		return nil, err
	}
	logger.Info("telegram bot authorized on account", zap.String("bot_name", user.UserName), zap.Int64("ID", user.ID))

	bot.Debug = true // TODO: брать из конфига

	return &TelegramBot{
		Api: bot,
	}, nil
}

func (b *TelegramBot) GetUpdates(timeout time.Duration) tgbotapi.UpdatesChannel {
	updates := b.Api.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout:        int(timeout.Seconds()),
		AllowedUpdates: []string{"message", "callback_query", "my_chat_member"},
	})
	return updates
}

func (b *TelegramBot) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		b.Api.StopReceivingUpdates()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return fmt.Errorf("bot shutdown timeout: %w", ctx.Err())
	}
	return nil
}

func (b *TelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	sent, err := b.Api.Send(c)
	return sent, err
}
