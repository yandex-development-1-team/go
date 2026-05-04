package repository

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/models"
)

// ErrSessionNotFound возвращается, если сессии пользователя в Redis нет.
var ErrSessionNotFound = errors.New("session not found")

// WithDBMetrics оборачивает операцию к Postgres метриками и нормализацией ошибок.
func WithDBMetrics(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	seconds := time.Since(start).Minutes()

	metrics.ObserveDatabaseQueryDuration(operation, seconds)
	if err != nil {
		metrics.IncDatabaseErrors(operation)
		return CheckDBError(operation, err)
	}

	return err
}

// WithDBMetricsValue то же, что WithDBMetrics, для функций с возвращаемым значением.
func WithDBMetricsValue[T any](operation string, fn func() (T, error)) (T, error) {
	start := time.Now()
	result, err := fn()
	seconds := time.Since(start).Minutes()

	metrics.ObserveDatabaseQueryDuration(operation, seconds)
	if err != nil {
		metrics.IncDatabaseErrors(operation)
		return result, CheckDBError(operation, err)
	}

	return result, err
}

// WithRedisMetrics оборачивает операцию к Redis метриками.
func WithRedisMetrics(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	seconds := time.Since(start).Seconds()

	metrics.ObserveCacheSetDuration(operation, seconds)
	if err != nil {
		metrics.IncCacheErrors(operation)
		logger.Error("redis_error", zap.Error(err), zap.String("operation", operation))
		return err
	}

	return err
}

// WithRedisMetricsValue то же, что WithRedisMetrics, для функций с возвращаемым значением.
func WithRedisMetricsValue[T any](operation string, fn func() (T, error)) (T, error) {
	start := time.Now()
	result, err := fn()
	seconds := time.Since(start).Seconds()

	metrics.ObserveCacheSetDuration(operation, seconds)
	if err != nil {
		metrics.IncCacheErrors(operation)
		if errors.Is(err, ErrSessionNotFound) {
			return result, err
		}
		logger.Error("redis_error", zap.Error(err), zap.String("operation", operation))
		return result, err
	}

	return result, err
}

// CheckDBError приводит сырую ошибку Postgres к доменным sentinel-ам.
func CheckDBError(operation string, err error) error {
	if errors.Is(err, models.ErrUnauthorized) {
		return models.ErrUnauthorized
	}
	if errors.Is(err, models.ErrForbidden) {
		return models.ErrForbidden
	}
	if errors.Is(err, models.ErrBookingNotFound) {
		logger.Info("booking_not_found", zap.Error(err), zap.String("operation", operation))
		return models.ErrBookingNotFound
	}
	if errors.Is(err, models.ErrSlotOccupied) {
		logger.Info("slot_is_already_occupied", zap.Error(err), zap.String("operation", operation))
		return models.ErrSlotOccupied
	}
	if errors.Is(err, models.ErrUserNotFound) {
		logger.Info("user_not_found", zap.Error(err), zap.String("operation", operation))
		return models.ErrUserNotFound
	}
	if errors.Is(err, models.ErrApplicationNotFound) {
		logger.Info("application_not_found", zap.Error(err), zap.String("operation", operation))
		return models.ErrApplicationNotFound
	}
	if errors.Is(err, context.Canceled) {
		logger.Info("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return models.ErrRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Info("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return models.ErrRequestTimeout
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "57014" {
		logger.Info("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return models.ErrRequestTimeout
	}
	logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return models.ErrDatabase
}
