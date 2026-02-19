package repository

import (
	"context"
	"errors"

	"github.com/yandex-development-1-team/go/internal/models"
)

func withMetricsValue[T any](operation string, repo func() (T, error)) (T, error) {
	//srart := time.Now()
	result, err := repo()
	//seconds := time.Since(srart).Minutes()

	if err != nil {
		//metrics.IncDatabaseErrors(operation)
		return result, checkError(operation, err)
	}
	//metrics.ObserveDatabaseQueryDuration(operation, seconds)

	return result, err
}

func withMetrics(operation string, repo func() error) error {
	//srart := time.Now()
	err := repo()
	//seconds := time.Since(srart).Minutes()

	//metrics.ObserveDatabaseQueryDuration(operation, seconds)
	if err != nil {
		//metrics.IncDatabaseErrors(operation)
		return checkError(operation, err)
	}

	return err
}

func checkError(operation string, err error) error {
	if errors.Is(err, models.ErrBookingNotFound) {
		//logger.Info("booking_not_found", zap.Error(err), zap.String("operation", operation))
		return models.ErrBookingNotFound
	}
	if errors.Is(err, models.ErrSlotOccupied) {
		//logger.Info("slot_is_already_occupied", zap.Error(err), zap.String("operation", operation))
		return models.ErrSlotOccupied
	}
	if errors.Is(err, models.ErrUserNotFound) {
		//logger.Info("user_not_found", zap.Error(err), zap.String("operation", operation))
		return models.ErrUserNotFound
	}
	if errors.Is(err, context.Canceled) {
		//logger.Info("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return models.ErrRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		//logger.Info("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return models.ErrRequestTimeout
	}

	//logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return models.ErrDatabase
}
