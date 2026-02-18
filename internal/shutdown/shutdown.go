package shutdown

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	var err error
	// канал для прослушивания
	shutdown := make(chan os.Signal, 1)
	// os.Interrupt для кроссплатформенности
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	// ждем сигнала завершения выполнения приложения
	signal := <-shutdown
	s.logger.Info("получен сигнал завершения работы приложения")

	// выбор длительности таймаута контекста при shutdown
	var timeout int
	switch signal {
	case os.Interrupt:
		// для ctrl+c
		timeout = 30
		s.logger.Info("slow shutdown started")
	case syscall.SIGTERM:
		// для команды из ОС
		timeout = 5
		s.logger.Info("slow shutdown started")
	}

	// обший таймаут на shutdown
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout))
	defer cancel()

	// закрытие канала входящих сообщений тг бота
	s.bot.StopReceivingUpdates()
	s.logger.Info("tg bot's incoming messages channel closed")

	// закрытие соединения с бд
	timeoutForDB := timeout * 6 / 10
	ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(timeoutForDB)*time.Second)
	defer cancel2()
	closed := make(chan error, 1)
	s.logger.Info("stoping database started")
	go func() {
		closed <- s.db.Close()
	}()

	select {
	// для корректного закрытия
	case err = <-closed:
		if err != nil {
			s.logger.Error(err.Error())
		}
	// для выхода по таймауту
	case <-ctx2.Done():
		s.logger.Error("database shutdown timeout")
		err = fmt.Errorf("database shutdown timeout: %w", ctx.Err())
	}
	if err == nil {
		s.logger.Info("database stopped successfully")
	}

	// закрытие HTTP сервера метрик
	timeoutForMetrics := (timeout - timeoutForDB) / 2
	ctx3, cancel3 := context.WithTimeout(ctx, time.Duration(timeoutForMetrics)*time.Second)
	defer cancel3()
	s.metrics.Shutdown(ctx3)
	s.logger.Info("prometheus metrics server stopped")

	// завершение работы логгера
	s.logger.Info("stoping logger")
	_ = s.logger.Sync()

	return err
}
