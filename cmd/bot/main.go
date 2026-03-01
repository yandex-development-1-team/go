package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yandex-development-1-team/go/internal/service"

	"github.com/yandex-development-1-team/go/internal/database"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/repository/mocks"

	"github.com/yandex-development-1-team/go/internal/api"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/shutdown"

	_ "github.com/lib/pq"
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

	// init database connection for migrations
	db, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// run migrations
	if err := database.RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	redisClient, err := sr.NewRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("init redis client: %w", err)
	}

	// init telegram bot
	tgBot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	// init signal ctx
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// init rate limiters
	apiRL := bot.NewApiRateLimiter(cfg.ApiRPS)
	msgRL, err := bot.NewMsgRateLimiter(cfg.CacheSizeRPS, cfg.MsgRPS)
	if err != nil {
		return fmt.Errorf("failed to init new messages rate limiter")
	}

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

	// run metrics server
	go func() {
		logger.Info("starting metrics and health server", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	rep := repository.NewRepository()
	repMock := mocks.NewMockClient()

	startHandler := handlers.NewStartHandler(tgBot, rep)
	bsService := service.NewBoxSolutionsService(repMock)
	boxSolutionsHandler := handlers.NewBoxSolutions(tgBot, bsService)

	// init handler
	handler := handlers.NewHandler(tgBot, msgRL, startHandler, boxSolutionsHandler)
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
