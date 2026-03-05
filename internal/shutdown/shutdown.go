package shutdown

import (
	"context"
	"errors"
	"sync"

	"github.com/yandex-development-1-team/go/internal/logger"
)

type DB interface {
	Close(ctx context.Context) error
}

type Bot interface {
	Shutdown(ctx context.Context) error
}

type MetricsServer interface {
	Shutdown(ctx context.Context) error
}

type ShutdownHandler struct {
	bot     Bot
	db      DB
	metrics MetricsServer
}

func NewShutdownHandler(bot Bot, db DB, metrics MetricsServer) *ShutdownHandler {
	return &ShutdownHandler{
		bot:     bot,
		db:      db,
		metrics: metrics,
	}
}

func (s *ShutdownHandler) WaitForShutdown(ctx context.Context) error {
	var wg sync.WaitGroup

	errChan := make(chan error, 3)

	// bot shutdown
	logger.Info("shutting down bot...")
	wg.Go(func() {
		if err := s.bot.Shutdown(ctx); err != nil {
			errChan <- err
		} else {
			logger.Info("bot gracefully shutdown")
		}
	})

	// DB shutdown
	// logger.Info("shutting down DB...")
	// wg.Go(func() {
	// 	if err := s.db.Close(ctx); err != nil {
	// 		errChan <- err
	// 	} else {
	// 		logger.Info("DB gracefully shutdown")
	// 	}
	// })

	// shutdown metrics server
	logger.Info("shutting down metrics server...")
	wg.Go(func() {
		if err := s.metrics.Shutdown(ctx); err != nil {
			errChan <- err
		} else {
			logger.Info("metrics server gracefully shutdown")
		}
	})

	wg.Wait()
	close(errChan)

	// collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
