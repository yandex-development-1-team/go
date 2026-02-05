package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/delivery/telegram"
	"go.uber.org/zap"
)

func main() {
	// load config
	cfg := config.MustLoadConfig()
	if err := run(cfg); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	// init logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// init telegram bot
	bot, err := bot.NewTelegramBot(cfg.TelegramBotToken, logger)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// get channel with updates
	updates := bot.GetUpdates(30 * time.Second)

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init handler
	handler := telegram.NewHandler(bot)

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
