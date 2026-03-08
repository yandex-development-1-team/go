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
	errChan := make(chan error, 3)

	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				errChan <- err
			}
		}()
	}

	// bot shutdown
	if s.bot != nil {
		logger.Info("shutting down bot...")
		run(func() error {
			err := s.bot.Shutdown(ctx)
			if err == nil {
				logger.Info("bot gracefully shutdown")
			}
			return err
		})
	}

	// redis shutdown
	logger.Info("shutting down redis...")
	run(func() error {
		err := s.redis.Shutdown(ctx).Err()
		if err == nil {
			logger.Info("redis gracefully shutdown")
		}
		return err
	})

	// DB shutdown
	logger.Info("shutting down DB...")
	run(func() error {
		err := s.db.Close()
		if err == nil {
			logger.Info("DB gracefully shutdown")
		}
		return err
	})

	// shutdown metrics server
	if s.metrics != nil {
		logger.Info("shutting down metrics server...")
		run(func() error {
			err := s.metrics.Shutdown(ctx)
			if err == nil {
				logger.Info("metrics server gracefully shutdown")
			}
			return err
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
