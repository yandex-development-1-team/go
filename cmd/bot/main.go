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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/api"
	"github.com/yandex-development-1-team/go/internal/api/server"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database"
	"github.com/yandex-development-1-team/go/internal/database/repository"
	apiHandlers "github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	apiRepository "github.com/yandex-development-1-team/go/internal/repository/postgres"
	"github.com/yandex-development-1-team/go/internal/service"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
	"github.com/yandex-development-1-team/go/internal/shutdown"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func run() error {
	// --- Config & logging ---
	cfg, err := config.GetConfig(nil)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	logger.NewLogger(cfg.Environment, cfg.LogLevel)
	defer logger.Sync()
	metrics.Initialize(cfg)

	// --- Infrastructure: DB (sqlx) ---
	db, err := sql.Open("postgres", cfg.DB.PostgresURL)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	if err := database.RunMigrations(db); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	dbSqlx := sqlx.NewDb(db, "postgres")

	// --- Infrastructure: Redis ---
	redisClient, err := repository.NewRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer redisClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	// --- Repositories ---
	boxSolutionRepo := repository.NewBoxSolutionRepo(dbSqlx)
	settingsRepo := apiRepository.NewSettingsRep(dbSqlx)
	specialProjectRepo := apiRepository.NewSpecialProjectRepository(dbSqlx)

	// --- Services ---
	settingsService := apiService.NewSettingsService(settingsRepo) // TODO: wire into API routes
	boxService := apiService.NewAPIBoxService(boxSolutionRepo)
	specialProjectService := service.NewSpecialProjectService(specialProjectRepo)

	// --- HTTP: metrics + health ---
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.NewHandler())
	metricsMux.HandleFunc("/health", api.NewHealthHandler(dbSqlx, cfg.TelegramBotAPIUrl))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.PrometheusPort),
		Handler: metricsMux,
	}
	go func() {
		logger.Info("metrics server starting", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server", zap.Error(err))
		}
	}()

	// --- API server (routers) ---
	apiServer := server.New(&cfg, &server.APIServices{
		BoxService:        boxService,
		SpecialProjectSvc: specialProjectService,
		SettingsService:   settingsService,
	})
	apiServer.RegisterRoutes()

	var wg sync.WaitGroup
	go func() {
		wg.Go(func() {
			if err := apiServer.Run(&cfg); err != nil {
				logger.Error("api server", zap.Error(err))
			}
			logger.Info("api server started", zap.Int("port", cfg.Port))
		})
	}()

	// --- API-only mode: no bot ---
	if cfg.APIOnly {
		logger.Info("api_only mode (no bot)")
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		return shutdown.NewShutdownHandler(nil, dbSqlx, metricsServer, redisClient).WaitForShutdown(shutdownCtx)
	}

	// --- Telegram bot ---
	tgBot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("telegram bot: %w", err)
	}
	apiRL := bot.NewApiRateLimiter(cfg.ApiRPS)
	msgRL, err := bot.NewMsgRateLimiter(cfg.CacheSizeRPS, cfg.MsgRPS)
	if err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	bsService := service.NewBoxSolutionsService(boxSolutionRepo)
	startHandler := apiHandlers.NewStartHandler(tgBot, nil)
	boxSolutionsHandler := apiHandlers.NewBoxSolutions(tgBot, bsService)
	handler := apiHandlers.NewHandler(tgBot, msgRL, startHandler, boxSolutionsHandler)
	logger.Info("bot started", zap.String("env", cfg.Environment))

	updates := tgBot.GetUpdates(30 * time.Second)
	go func() {
		for update := range updates {
			wg.Go(func() {
				if err := apiRL.Exec(ctx, func() { handler.Handle(ctx, update) }); err != nil {
					logger.Error("handle update", zap.Error(err))
				}
			})
		}
	}()

	// --- Shutdown ---
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := shutdown.NewShutdownHandler(tgBot, dbSqlx, metricsServer, redisClient).WaitForShutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info("waiting for in-flight updates...")
	updatesDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(updatesDone)
	}()
	select {
	case <-updatesDone:
		logger.Info("updates done")
	case <-shutdownCtx.Done():
		logger.Warn("shutdown timeout waiting for updates")
	}

	return nil
}
