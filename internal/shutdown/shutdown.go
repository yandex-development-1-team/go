package shutdown

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/yandex-development-1-team/go/internal/logger"
)

type DB interface {
	Close() error
}

type Bot interface {
	Shutdown(ctx context.Context) error
}

type MetricsServer interface {
	Shutdown(ctx context.Context) error
}

type Redis interface {
	Shutdown(ctx context.Context) *redis.StatusCmd
}

type ShutdownHandler struct {
	bot     Bot
	db      DB
	metrics MetricsServer
	redis   Redis
}

func NewShutdownHandler(bot Bot, db DB, metrics MetricsServer, redis Redis) *ShutdownHandler {
	return &ShutdownHandler{
		bot:     bot,
		db:      db,
		metrics: metrics,
		redis:   redis,
	}
}

func (s *ShutdownHandler) WaitForShutdown(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	if s.bot != nil {
		logger.Info("shutting down bot...")
		wg.Go(func() {
			if err := s.bot.Shutdown(ctx); err != nil {
				errChan <- err
			} else {
				logger.Info("bot gracefully shutdown")
			}
		})
	}

	if s.redis != nil {
		logger.Info("shutting down redis...")
		wg.Go(func() {
			if err := s.redis.Shutdown(ctx).Err(); err != nil {
				errChan <- err
			} else {
				logger.Info("redis gracefully shutdown")
			}
		})
	}

	if s.db != nil {
		logger.Info("shutting down DB...")
		wg.Go(func() {
			if err := s.db.Close(); err != nil {
				errChan <- err
			} else {
				logger.Info("DB gracefully shutdown")
			}
		})
	}

	if s.metrics != nil {
		logger.Info("shutting down metrics server...")
		wg.Go(func() {
			if err := s.metrics.Shutdown(ctx); err != nil {
				errChan <- err
			} else {
				logger.Info("metrics server gracefully shutdown")
			}
		})
	}

	wg.Wait()
	close(errChan)

	// collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
