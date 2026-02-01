package shutdown

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type ShutdownInterface interface {
	WaitForShutdown(context.Context) error
}

type ShutdownHandler struct {
	logger  *zap.Logger
	db      *sqlx.DB
	bot     *tgbotapi.BotAPI
	metrics *http.Server
}

func NewShutdownHandler(logger *zap.Logger, db *sqlx.DB, bot *tgbotapi.BotAPI, metrics *http.Server) *ShutdownHandler {
	return &ShutdownHandler{
		logger:  logger,
		db:      db,
		bot:     bot,
		metrics: metrics,
	}
}

func (s *ShutdownHandler) WaitForShutdown(ctx context.Context) error {
	// канал для прослушивания
	shutdown := make(chan os.Signal, 1)
	// os.Interrupt для кроссплатформенности
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	// ждем сигнала завершения выполнения приложения
	signal := <-shutdown
	s.logger.Info("получен сигнал завершения работы приложения")

	// выбор длительности таймаута контекста при shutdown
	var err error
	var timeout int
	switch signal {
	case os.Interrupt:
		// для ctrl+c
		timeout = 30
		s.logger.Info("slow shutdown started")
		err = s.slowShutdown(ctx, timeout)
	case syscall.SIGTERM:
		// для команды из ОС
		timeout = 5
		s.logger.Info("slow shutdown started")
		err = s.fastShutdown(ctx, timeout)
	}

	return err
}

func (s *ShutdownHandler) fastShutdown(ctx context.Context, timeout int) error {
	var err error

	// обший таймаут на shutdown
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout))
	defer cancel()

	// закрытие канала входящих сообщений тг бота
	s.bot.StopReceivingUpdates()
	s.logger.Info("tg bot's incoming messages channel closed")

	// закрытие соединения с бд
	// отводится 2 секунды или жесткое закрытие соединений и запросов
	ctx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()
	closed := make(chan error, 1)
	s.logger.Info("stoping database started")
	go func() {
		closed <- s.db.Close()
	}()

	select {
	// для быстрого корректного закрытия
	case err = <-closed:
		if err != nil {
			_ = terminateAllQueries(s.db)
			s.logger.Error(err.Error())
		}
	// для закрытия по таймауту и жесткого закрытия всех соединений с бд и запросов
	case <-ctx2.Done():
		err = terminateAllQueries(s.db)
		if err != nil {
			s.logger.Error(err.Error())
		}
	}

	if err == nil {
		s.logger.Info("database stopped successfully")
	}

	// закрытие HTTP сервера метрик
	ctx3, cancel3 := context.WithTimeout(ctx, 1*time.Second)
	defer cancel3()
	s.metrics.Shutdown(ctx3)
	s.logger.Info("prometheus metrics server stopped")

	return err
}

func (s *ShutdownHandler) slowShutdown(ctx context.Context, timeout int) error {
	var err error

	// общий таймаут на shutdown
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout))
	defer cancel()

	// закрытие канала входящих сообщений тг бота
	s.bot.StopReceivingUpdates()
	s.logger.Info("tg bot's incoming messages channel closed")

	// закрытие соединения с бд
	s.logger.Info("stoping database started")
	err = shutdownDBConn(ctx, s.db)
	if err != nil {
		s.logger.Error(err.Error())
	} else {
		s.logger.Info("database stopped successfully")
	}

	// закрытие HTTP сервера метрик
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	s.logger.Info("start stopping prometheus metrics server")
	s.metrics.Shutdown(ctx2)
	s.logger.Info("prometheus metrics server stopped")

	return err
}

func shutdownDBConn(ctx context.Context, db *sqlx.DB) error {
	ctx, cancel20 := context.WithTimeout(ctx, 20)
	defer cancel20()

	// канал получения сингала закрытия БД
	done := make(chan error)
	defer close(done)
	go func() {
		done <- db.Close()
	}()

	select {
	// корректное закрытие соединений за 20 секунд
	case err := <-done:
		if err != nil {
			// если закрытие бд завершилось ошибкой - убиваем запросы
			_ = terminateAllQueries(db)
			return fmt.Errorf("failed to close database: %v", err)
		}
	case <-ctx.Done():
		// не завершились за 20 секунд - зависшие или долгие запросы
		// убиваем запросы и фиксируем результат
		err := terminateAllQueries(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func terminateAllQueries(db *sqlx.DB) error {
	_, err := db.Exec(`
			SELECT pg_terminate_backend(pid) 
			FROM pg_stat_activity 
			WHERE usename = CURRENT_USER 
				AND pid != pg_backend_pid()
				AND state = 'active'
	`)
	return err
}
