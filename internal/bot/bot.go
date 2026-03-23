package bot

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"

	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/logger"
)

type TelegramBot struct {
	Api *tgbotapi.BotAPI
}

func NewTelegramBot(cfg config.Telegram) (*TelegramBot, error) {
	var bot *tgbotapi.BotAPI

	if cfg.Proxy.Enabled && cfg.Proxy.Server != "" {

		dialer, err := proxy.SOCKS5("tcp",
			fmt.Sprintf("%s:%s", cfg.Proxy.Server, cfg.Proxy.Port),
			nil,
			proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy dialer: %w", err)
		}

		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				},
				TLSHandshakeTimeout: 10 * time.Second,
			},
			Timeout: 60 * time.Second,
		}
		apiEndPoint := fmt.Sprintf("%s%s", cfg.ApiUrl, "/bot%s/%s")
		bot, err = tgbotapi.NewBotAPIWithClient(cfg.BotToken, apiEndPoint, client)
		if err != nil {
			return nil, fmt.Errorf("failed to create bot with proxy: %w", err)
		}
		logger.Info("telegram bot initialized with proxy",
			zap.String("server", cfg.Proxy.Server),
			zap.String("port", cfg.Proxy.Port))
	} else {
		var err error
		bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
		if err != nil {
			return nil, err
		}

		user, err := bot.GetMe()
		if err != nil {
			return nil, err
		}
		logger.Info("telegram bot authorized on account", zap.String("bot_name", user.UserName), zap.Int64("ID", user.ID))

		bot.Debug = cfg.Debug

	}

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
