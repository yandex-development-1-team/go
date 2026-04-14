package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/yandex-development-1-team/go/internal/models"
)

type ObjectStorage interface {
	UploadFile(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error)
	RemoveFile(ctx context.Context, objectName string) error
}

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByUUID(ctx context.Context, fileUUID uuid.UUID) (*models.File, error)
	GetByURL(ctx context.Context, url string) (*models.File, error)
	DeactivateByURL(ctx context.Context, url string) error
	ListInactiveOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error)
	IsFileReferenced(ctx context.Context, file models.File) (bool, error)
	DeleteHard(ctx context.Context, fileID int64) error
}
