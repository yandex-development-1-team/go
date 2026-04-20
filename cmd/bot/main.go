package main

import (
	"context"
	"database/sql"
	"errors"
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
	botHandlers "github.com/yandex-development-1-team/go/internal/handlers"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/repository/postgres"
	"github.com/yandex-development-1-team/go/internal/repository/redis"
	"github.com/yandex-development-1-team/go/internal/service"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
	"github.com/yandex-development-1-team/go/internal/shutdown"
	minioStorage "github.com/yandex-development-1-team/go/internal/storage/minio"
	"github.com/yandex-development-1-team/go/internal/worker"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func run() error {
	cfg, err := config.GetConfig(nil)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	logger.NewLogger(cfg.Environment, cfg.LogLevel)
	defer func() { _ = logger.Sync() }()
	metrics.Initialize(cfg)

	fmt.Println(cfg.DB.PostgresURL)
	db, err := sql.Open("postgres", cfg.DB.PostgresURL)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	migrationsDir, err := database.ResolveMigrationsDir(cfg.MigrationsDir)
	if err != nil {
		return fmt.Errorf("migrations dir: %w", err)
	}
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	dbSqlx := sqlx.NewDb(db, "postgres")

	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer func() { _ = redisClient.Close() }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fileStorage, err := minioStorage.New(cfg.Storage)
	if err != nil {
		return fmt.Errorf("init minio storage: %w", err)
	}

	if err := fileStorage.EnsureBucket(ctx); err != nil {
		return fmt.Errorf("ensure minio bucket: %w", err)
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("redis connected", zap.String("addr", cfg.Redis.Addr))

	telegramUserRepo := postgres.NewTelegramUserRepository(dbSqlx)
	boxSolutionRepo := postgres.NewBoxSolutionRepo(dbSqlx)
	bookRepo := postgres.NewBookingRepository(dbSqlx)
	sessionRepo := redis.NewSessionRepository(redisClient, redis.WithTTL(cfg.Session.TTL))
	settingsRepo := postgres.NewSettingsRep(dbSqlx)
	specialProjectRepo := postgres.NewSpecialProjectRepository(dbSqlx)
	refreshTokenRepoRepo := postgres.NewRefreshTokenRepo(dbSqlx)
	txRepo := postgres.NewTxRepo(dbSqlx)
	staffRepo := postgres.NewStaffRepo(dbSqlx)
	analyticsRepo := postgres.NewAnalyticsRepo(dbSqlx)
	resourcePageRepo := postgres.NewResourcePageRepo(dbSqlx)
	passwordResetRepo := postgres.NewPasswordResetRepository(dbSqlx)
	emailService := apiService.NewEmailService(cfg.Email)
	fileRepo := postgres.NewFileRepository(dbSqlx)
	applicationRepo := postgres.NewApplicationRepository(dbSqlx)

	settingsService := apiService.NewSettingsService(settingsRepo)
	bookService := botService.NewBookingService(sessionRepo, bookRepo, boxSolutionRepo)
	keyboard := botHandlers.NewKeyboardService()
	bsService := service.NewBoxSolutionsService(boxSolutionRepo)
	detailService := botService.NewDetailService(boxSolutionRepo)
	fileService := apiService.NewFileService(fileRepo, fileStorage)
	aboutService := botService.NewAboutService(resourcePageRepo)
	guideService := botService.NewGuideService(resourcePageRepo)
	exampleService := botService.NewExamplesSpService(resourcePageRepo)
	linksService := botService.NewUsefulLinksService(resourcePageRepo)
	reqSpService := botService.NewRequestSpService(resourcePageRepo)
	boxService := apiService.NewAPIBoxService(boxSolutionRepo, fileService)
	specialProjectService := service.NewSpecialProjectService(specialProjectRepo)
	analyticsService := apiService.NewAnalyticsService(analyticsRepo)
	resourcePageService := service.NewResourcePageService(resourcePageRepo)
	userService := apiService.NewUserService(staffRepo)
	applicationSvc := apiService.NewApplicationsService(applicationRepo, txRepo)
	bookAPISvc := apiService.NewBookingsService(bookRepo, txRepo)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.NewHandler())
	metricsMux.HandleFunc("/health", api.NewHealthHandler(dbSqlx, cfg.TelegramBotAPIUrl))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.PrometheusPort),
		Handler: metricsMux,
	}
	go func() {
		logger.Info("metrics server starting", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server", zap.Error(err))
		}
	}()

	apiAuthService := apiService.NewAuthService(dbSqlx, refreshTokenRepoRepo, passwordResetRepo, staffRepo, emailService, txRepo, cfg.AuthConfig.JWTSecret,
		cfg.AuthConfig.AccessTokenTTLMinutes, cfg.AuthConfig.RefreshTokenTTLDays)

	apiServer := server.New(&cfg, &server.APIServices{
		BoxService:        boxService,
		SpecialProjectSvc: specialProjectService,
		SettingsService:   settingsService,
		AnalyticsSvc:      analyticsService,
		RecPageSvc:        resourcePageService,
		UserSvc:           userService,
		FileService:       fileService,
		ApplicationRepo:   applicationRepo,
		MiddlewareRepo:    dbSqlx,
		ApplicationSvc:    applicationSvc,
		BookingSvc:        bookAPISvc,
	}, apiAuthService)

	if cfg.FileGC.Enabled {
		fileCleanupWorker := worker.NewFileCleanupWorker(
			fileService,
			cfg.FileGC.Interval,
			cfg.FileGC.OrphanGrace,
			cfg.FileGC.DeleteBatchSize,
		)

		go fileCleanupWorker.Start(ctx)
		logger.Info("file cleanup worker started",
			zap.Duration("interval", cfg.FileGC.Interval),
			zap.Duration("orphan_grace_period", cfg.FileGC.OrphanGrace),
			zap.Int("delete_batch_size", cfg.FileGC.DeleteBatchSize),
		)
	}

	apiServer.RegisterRoutes(cfg.YandexForms.WebhookToken)

	var wg sync.WaitGroup
	go func() {
		wg.Go(func() {
			if err := apiServer.Run(&cfg); err != nil {
				logger.Error("api server", zap.Error(err))
			}
			logger.Info("api server started", zap.Int("port", cfg.Port))
		})
	}()

	if cfg.APIOnly {
		logger.Info("api_only mode (no bot)")
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		return shutdown.NewShutdownHandler(nil, dbSqlx, metricsServer, redisClient).WaitForShutdown(shutdownCtx)
	}

	tgBot, err := bot.NewTelegramBot(cfg.Telegram)
	if err != nil {
		return fmt.Errorf("telegram bot: %w", err)
	}
	apiRL := bot.NewApiRateLimiter(cfg.ApiRPS)
	msgRL, err := bot.NewMsgRateLimiter(cfg.CacheSizeRPS, cfg.MsgRPS)
	if err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	startHandler := botHandlers.NewStartHandler(tgBot.Api, telegramUserRepo, sessionRepo)
	statusHandler := botHandlers.NewStatusHandler(tgBot.Api, bookRepo, sessionRepo)
	bsHandler := botHandlers.NewBoxSolutions(tgBot.Api, bsService)
	bcHandler := botHandlers.NewBookingFormHandler(tgBot.Api, bookService, startHandler, bsHandler, keyboard)
	infoHandler := botHandlers.NewDetailHandler(detailService, tgBot.Api, startHandler, bsHandler, keyboard)

	aboutHandler := botHandlers.NewAboutHandler(aboutService, tgBot.Api, startHandler, bsHandler, keyboard)
	guideHandler := botHandlers.NewGuideHandler(guideService, tgBot.Api, startHandler, bsHandler, keyboard)
	exampleHandler := botHandlers.NewExamplesSpHandler(exampleService, tgBot.Api, startHandler, bsHandler, keyboard)
	linksHandler := botHandlers.NewUsefulLinksHandler(linksService, tgBot.Api, startHandler, bsHandler, keyboard)
	reqSpHandler := botHandlers.NewRequestSpHandler(reqSpService, tgBot.Api, startHandler, bsHandler, keyboard)

	callbackRouter := botHandlers.NewCallbackRouter(tgBot.Api)
	msgRouter := botHandlers.NewMessageRouter(tgBot.Api, startHandler, statusHandler, sessionRepo, bcHandler, msgRL)

	callbackRouter.Register(botHandlers.CallbackBoxSolutions, bsHandler)
	callbackRouter.Register(botService.CallbackBookingPrefix, bcHandler)
	callbackRouter.Register(botHandlers.CallbackInfoPrefix, infoHandler)
	callbackRouter.Register(botHandlers.BoxSolutionsButtonBackToMainMenu, startHandler)

	callbackRouter.Register(botHandlers.CallbackAboutUs, aboutHandler)
	callbackRouter.Register(botHandlers.CallbackVisitGuide, guideHandler)
	callbackRouter.Register(botHandlers.CallbackProjectExamples, exampleHandler)
	callbackRouter.Register(botHandlers.CallbackSupport, linksHandler)
	callbackRouter.Register(botHandlers.CallbackSpecialProject, reqSpHandler)

	handler := botHandlers.NewHandler(tgBot, msgRL, msgRouter, callbackRouter)

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
