package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
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

	// init telegram tgBot
	tgBot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// get channel with updates
	updates := tgBot.GetUpdates(30 * time.Second)

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init rate limiters
	apiRL := bot.NewApiRateLimiter()
	msgRL := bot.NewMsgRateLimiter()

	// init handler
	handler := handlers.NewHandler(tgBot, msgRL)

	logger.Info("bot has been started",
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.LogLevel),
	)

	// handle updates
	var wg sync.WaitGroup
	go func() {
		for update := range updates {
			wg.Go(func() {
				if err := apiRL.Exec(ctx, func() { handler.Handle(ctx, update) }); err != nil {
					logger.Error("failed to handle update", zap.Error(err))
				}
			})
		}
	}()

	<-ctx.Done()

	// wait for updates to finish
	logger.Info("waiting for updates to finish...")
	updatesDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(updatesDone)
	}()

	select {
	case <-updatesDone:
		logger.Info("all updates processed")
	case <-time.After(30 * time.Second):
		logger.Warn("timeout waiting for updates to finish")
	}

	// bot gracefully shutdown
	logger.Info("shutting down bot...")
	botDone := make(chan struct{})
	go func() {
		tgBot.Shutdown()
		close(botDone)
	}()

	select {
	case <-botDone:
		logger.Info("bot gracefully shutdown")
	case <-time.After(30 * time.Second):
		logger.Warn("bot timeout shutdown")
	}

	return nil
}
