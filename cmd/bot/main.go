package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yandex-development-1-team/go/internal/api"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/shutdown"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error:%v\n", err)
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

	// init telegram tgBot
	tgBot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init rate limiters
	apiRL := bot.NewApiRateLimiter(cfg.ApiRPS)
	msgRL := bot.NewMsgRateLimiter(cfg.MsgRPS)

	// init metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.NewHandler())
	metricsMux.HandleFunc("/health", api.NewHealthHandler(nil, cfg.TelegramBotAPIUrl))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.PrometheusPort),
		Handler: metricsMux,
	}

	// TODO: init API server
	/*apiMux := http.NewServeMux()
	// apiMux.HandleFunc("/", handlers.APIHandler)
	apiServer := &http.Server{
		Addr:    ":8080",
		Handler: apiMux,
	}*/

	// run metrics server
	go func() {
		logger.Info("starting metrics and health server", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	// init handler
	handler := handlers.NewHandler(tgBot, msgRL)
	logger.Info("bot has been started",
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.LogLevel),
	)

	// get channel with updates
	updates := tgBot.GetUpdates(30 * time.Second) // TODO: перенести в конфиг

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

	// wait for shutdown signal
	<-ctx.Done()

	// init ctx timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// shutdown bot and servers (this closes updates channel)
	shutdown := shutdown.NewShutdownHandler(tgBot, nil, metricsServer)
	if err := shutdown.WaitForShutdown(ctxTimeout); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	// wait for in-flight updates to finish
	logger.Info("waiting for updates to finish...")
	updatesDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(updatesDone)
	}()

	select {
	case <-updatesDone:
		logger.Info("all updates processed")
	case <-ctxTimeout.Done():
		logger.Warn("timeout waiting for updates to finish")
	}

	return nil
}
