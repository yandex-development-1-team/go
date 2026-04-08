package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

// FileCleanupWorker periodically removes inactive orphaned files.
type FileCleanupWorker struct {
	fileService     *apiService.FileService
	interval        time.Duration
	orphanGrace     time.Duration
	deleteBatchSize int
}

// NewFileCleanupWorker creates a new FileCleanupWorker.
func NewFileCleanupWorker(
	fileService *apiService.FileService,
	interval time.Duration,
	orphanGrace time.Duration,
	deleteBatchSize int,
) *FileCleanupWorker {
	return &FileCleanupWorker{
		fileService:     fileService,
		interval:        interval,
		orphanGrace:     orphanGrace,
		deleteBatchSize: deleteBatchSize,
	}
}

// Start runs the cleanup loop until the context is cancelled.
func (w *FileCleanupWorker) Start(ctx context.Context) {
	if w.fileService == nil {
		logger.Warn("file cleanup worker disabled: file service is nil")
		return
	}

	if w.interval <= 0 {
		logger.Warn("file cleanup worker disabled: interval <= 0")
		return
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("file cleanup worker stopped")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *FileCleanupWorker) runOnce(ctx context.Context) {
	olderThan := time.Now().UTC().Add(-w.orphanGrace)

	deletedCount, deletedBytes, err := w.fileService.CleanupInactiveFiles(ctx, olderThan, w.deleteBatchSize)
	if err != nil {
		logger.Error("file cleanup failed", zap.Error(err))
		return
	}

	if deletedCount == 0 {
		logger.Debug("file cleanup finished: nothing to delete")
		return
	}

	logger.Info(
		"file cleanup finished",
		zap.Int("deleted_count", deletedCount),
		zap.Int64("deleted_bytes", deletedBytes),
	)
}
