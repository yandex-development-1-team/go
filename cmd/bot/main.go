package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		logger.Fatal("error", zap.Error(err))
	}
}

func run() error {
	// init config
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	// init logger
	logger.NewLogger("dev", "debug")
	defer logger.Sync()

	// init telegram bot
	bot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// get channel with updates
	updates := bot.GetUpdates(30 * time.Second)

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init handler
	handler := handlers.NewHandler(bot)

	// handle updates
	go func() {
		for update := range updates {
			handler.Handle(ctx, update)
		}
	}()

	<-ctx.Done()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch := make(chan struct{})
	go func() {
		bot.Shutdown(ctxTimeout)
		close(ch)
	}()

	select {
	case <-ch:
		logger.Info("gracefully shutdown")
	case <-ctxTimeout.Done():
		logger.Warn("timeout graceful shutdown")
	}

	return nil
}
