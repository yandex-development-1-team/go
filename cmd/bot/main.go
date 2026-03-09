package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/yandex-development-1-team/go/internal/api/server"
	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/database/repository/mocks"
	postgresRepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
	"github.com/yandex-development-1-team/go/internal/service"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"

	"github.com/yandex-development-1-team/go/internal/api"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database"
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
	apiOnly := os.Getenv("RUN_MODE") == "api_only"

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
	db, err := sql.Open("postgres", cfg.DB.PostgresURL)
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
	// Оборачиваем sql.DB в sqlx.DB
	dbSqlx := sqlx.NewDb(db, "postgres")

	redisClient, err := repository.NewRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("init redis client: %w", err)
	}
	defer redisClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.NewHandler())
	metricsMux.HandleFunc("/health", api.NewHealthHandler(nil, cfg.TelegramBotAPIUrl))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.PrometheusPort),
		Handler: metricsMux,
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	go func() {
		logger.Info("starting metrics and health server", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	if apiOnly {
		logger.Info("running in api_only mode (no bot)")
		<-ctx.Done()
		ctxTimeout, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelShutdown()
		return shutdown.NewShutdownHandler(nil, dbSqlx, metricsServer, redisClient, nil).WaitForShutdown(ctxTimeout)
	}

	tgBot, err := bot.NewTelegramBot(cfg.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("failed to init telegram bot: %w", err)
	}

	apiRL := bot.NewApiRateLimiter(cfg.ApiRPS)
	msgRL, err := bot.NewMsgRateLimiter(cfg.CacheSizeRPS, cfg.MsgRPS)
	if err != nil {
		return fmt.Errorf("failed to init new messages rate limiter")
	}

	rep := repository.NewRepository(dbSqlx)
	repMock := mocks.NewMockClient(cfg.MockLocalDir)

	var bsService *service.BoxSolutionsService
	if cfg.MockClientEnabled {
		bsService = service.NewBoxSolutionsService(repMock)
	} else {
		bsService = service.NewBoxSolutionsService(rep)
	}

	startHandler := handlers.NewStartHandler(tgBot, rep)
	boxSolutionsHandler := handlers.NewBoxSolutions(tgBot, bsService)

	// init handler
	handler := handlers.NewHandler(tgBot, msgRL, startHandler, boxSolutionsHandler)
	logger.Info("bot has been started",
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.LogLevel),
	)

	apiBoxService := apiService.NewAPIBoxService()
	specProjRepo := postgresRepo.NewSpecialProjectRepository(dbSqlx)
	specProjSvc := service.NewSpecialProjectService(specProjRepo)

	var wg errgroup.Group

	// Creating an API server
	apiServer := server.New(&cfg)

	// Registering routes
	apiServer.RegisterRoutes(&server.APIServices{
		BoxService:        apiBoxService,
		SpecialProjectSvc: specProjSvc,
	})

	// Launching API server
	wg.Go(func() error {
		if err := apiServer.Run(&cfg); err != nil {
			logger.Error("failed to start the server", zap.Error(err))
			return err
		}
		logger.Info("server has been started", zap.Int("port", cfg.Port))
		return nil
	})

	// get channel with updates
	updates := tgBot.GetUpdates(cfg.GetUpdatesTimeout)

	// handle updates
	go func() {
		for update := range updates {
			u := update
			wg.Go(func() error {
				if err := apiRL.Exec(ctx, func() { handler.Handle(ctx, u) }); err != nil {
					logger.Error("failed to handle update", zap.Error(err))
					return err
				}
				return nil
			})
		}
	}()

	// wait for shutdown signal
	<-ctx.Done()

	// init ctx timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// shutdown bot, API server, and infra (this closes updates channel)
	shutdownHandler := shutdown.NewShutdownHandler(tgBot, dbSqlx, metricsServer, redisClient, apiServer)
	if err := shutdownHandler.WaitForShutdown(ctxTimeout); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	// wait for in-flight updates to finish
	logger.Info("waiting for updates to finish...")
	if err := wg.Wait(); err != nil {
		logger.Warn("some goroutines finished with error", zap.Error(err))
	}
	logger.Info("all updates processed")

	return nil
}
