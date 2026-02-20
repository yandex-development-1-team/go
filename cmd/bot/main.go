package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yandex-development-1-team/go/internal/api"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	sr "github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error:%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// init config
	cfg, err := config.GetConfig(nil)
	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	// init logger
	logger.NewLogger(cfg.Environment, cfg.LogLevel)
	defer logger.Sync()

	// init metrics
	metrics.Initialize(cfg)

	redisClient, err := sr.NewRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("init redis client: %w", err)
	}

	// init telegram bot
	bot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// init repos
	sessionRepo := sr.NewSessionRepository(
		redisClient,
		sr.WithTTL(cfg.Session.TTL),
	)

	// get channel with updates
	updates := bot.GetUpdates(30 * time.Second) // TODO: перенести в конфиг

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.NewHandler())
	metricsMux.HandleFunc("/health", api.NewHealthHandler(nil, cfg.TelegramBotAPIUrl))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.PrometheusPort),
		Handler: metricsMux,
	}

	// Redis connection check — fail fast.
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	// TODO: init API server
	/*apiMux := http.NewServeMux()
	// apiMux.HandleFunc("/", handlers.APIHandler)
	apiServer := &http.Server{
		Addr:    ":8080",
		Handler: apiMux,
	}*/

	var wg sync.WaitGroup

	// run metrics server
	wg.Go(func() {
		logger.Info("starting metrics and health server", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	})

	// init handler
	handler := handlers.NewHandler(bot)

	// handle updates
	go func() {
		for update := range updates {
			handler.Handle(ctx, update) // TODO: ассинхронный вызов
		}
	}()

	// wait for shutdown signal
	<-ctx.Done()
	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // TODO: вынести в конфиг ???
	defer cancel()

	ch := make(chan struct{})
	go func() {
		bot.Shutdown(shutdownCtx)
		close(ch)
	}()

	// shutdown servers concurrently
	errCh := make(chan error, 1) // metrics errors + TODO: API server

	wg.Go(func() {
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			errCh <- fmt.Errorf("metrics server shutdown error: %w", err)
		} else {
			errCh <- nil
		}
	})

	// collect shutdown errors
	shutdownErrors := 0
	for i := 0; i < 1; i++ {
		if err := <-errCh; err != nil {
			logger.Error("server shutdown error", zap.Error(err))
			shutdownErrors++
		}
	}

	// wait for all goroutines to finish
	wg.Wait()

	if shutdownErrors > 0 {
		return fmt.Errorf("encountered %d shutdown errors", shutdownErrors)
	}

	logger.Info("servers exited gracefully")
	return nil
}
