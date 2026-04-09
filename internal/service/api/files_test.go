package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yandex-development-1-team/go/internal/models"
)

type objectStorageMock struct {
	uploadFn func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error)
	removeFn func(ctx context.Context, objectName string) error
}

func (m *objectStorageMock) UploadFile(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
	return m.uploadFn(ctx, reader, objectName, size, contentType)
}

func (m *objectStorageMock) RemoveFile(ctx context.Context, objectName string) error {
	return m.removeFn(ctx, objectName)
}

type fileRepositoryMock struct {
	createFn                func(ctx context.Context, file *models.File) error
	getByUUIDFn             func(ctx context.Context, fileUUID uuid.UUID) (*models.File, error)
	getByURLFn              func(ctx context.Context, url string) (*models.File, error)
	deactivateByURLFn       func(ctx context.Context, url string) error
	listInactiveOlderThanFn func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error)
	isFileReferencedFn      func(ctx context.Context, file models.File) (bool, error)
	deleteHardFn            func(ctx context.Context, fileID int64) error
}

func (m *fileRepositoryMock) Create(ctx context.Context, file *models.File) error {
	return m.createFn(ctx, file)
}

func (m *fileRepositoryMock) GetByUUID(ctx context.Context, fileUUID uuid.UUID) (*models.File, error) {
	if m.getByUUIDFn == nil {
		return nil, nil
	}
	return m.getByUUIDFn(ctx, fileUUID)
}

func (m *fileRepositoryMock) GetByURL(ctx context.Context, url string) (*models.File, error) {
	if m.getByURLFn == nil {
		return nil, nil
	}
	return m.getByURLFn(ctx, url)
}

func (m *fileRepositoryMock) DeactivateByURL(ctx context.Context, url string) error {
	return m.deactivateByURLFn(ctx, url)
}

func (m *fileRepositoryMock) ListInactiveOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
	return m.listInactiveOlderThanFn(ctx, olderThan, limit)
}

func (m *fileRepositoryMock) IsFileReferenced(ctx context.Context, file models.File) (bool, error) {
	return m.isFileReferencedFn(ctx, file)
}

func (m *fileRepositoryMock) DeleteHard(ctx context.Context, fileID int64) error {
	return m.deleteHardFn(ctx, fileID)
}

func TestFileService_Upload_Success(t *testing.T) {
	repo := &fileRepositoryMock{
		createFn: func(ctx context.Context, file *models.File) error {
			if file.ObjectName == "" {
				t.Fatalf("expected object name to be set")
			}
			if file.URL == "" {
				t.Fatalf("expected URL to be set")
			}
			if file.OriginalName != "avatar.png" {
				t.Fatalf("unexpected original name: %s", file.OriginalName)
			}
			if !file.IsActive {
				t.Fatalf("expected file to be active")
			}
			return nil
		},
		deactivateByURLFn: func(ctx context.Context, url string) error { return nil },
		listInactiveOlderThanFn: func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
			return nil, nil
		},
		isFileReferencedFn: func(ctx context.Context, file models.File) (bool, error) { return false, nil },
		deleteHardFn:       func(ctx context.Context, fileID int64) error { return nil },
	}

	storage := &objectStorageMock{
		uploadFn: func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
			if objectName == "" {
				t.Fatalf("expected objectName")
			}
			if size <= 0 {
				t.Fatalf("expected positive size")
			}
			if contentType != "image/png" {
				t.Fatalf("unexpected content type: %s", contentType)
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("read uploaded payload: %v", err)
			}
			if string(body) != "hello-image" {
				t.Fatalf("unexpected file payload: %s", string(body))
			}
			return "http://localhost:9000/uploads/" + objectName, nil
		},
		removeFn: func(ctx context.Context, objectName string) error { return nil },
	}

	service := NewFileService(repo, storage)

	resp, err := service.Upload(
		context.Background(),
		bytes.NewBufferString("hello-image"),
		"avatar.png",
		"image/png",
		int64(len("hello-image")),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.UUID == "" {
		t.Fatalf("expected uuid in response")
	}
	if resp.URL == "" {
		t.Fatalf("expected url in response")
	}
	if resp.Filename != "avatar.png" {
		t.Fatalf("unexpected filename: %s", resp.Filename)
	}
}

func TestFileService_CleanupInactiveFiles_RemovesOrphaned(t *testing.T) {
	repo := &fileRepositoryMock{
		createFn:          func(ctx context.Context, file *models.File) error { return nil },
		deactivateByURLFn: func(ctx context.Context, url string) error { return nil },
		listInactiveOlderThanFn: func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
			return []models.File{
				{
					ID:         10,
					ObjectName: "old-file.png",
					URL:        "http://localhost:9000/uploads/old-file.png",
					SizeBytes:  128,
					IsActive:   false,
				},
			}, nil
		},
		isFileReferencedFn: func(ctx context.Context, file models.File) (bool, error) {
			return false, nil
		},
		deleteHardFn: func(ctx context.Context, fileID int64) error {
			if fileID != 10 {
				t.Fatalf("unexpected fileID: %d", fileID)
			}
			return nil
		},
	}

	removed := false
	storage := &objectStorageMock{
		uploadFn: func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "", nil
		},
		removeFn: func(ctx context.Context, objectName string) error {
			if objectName != "old-file.png" {
				t.Fatalf("unexpected object name: %s", objectName)
			}
			removed = true
			return nil
		},
	}

	service := NewFileService(repo, storage)

	count, bytesDeleted, err := service.CleanupInactiveFiles(context.Background(), time.Now(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !removed {
		t.Fatalf("expected file to be removed from storage")
	}
	if count != 1 {
		t.Fatalf("expected 1 deleted file, got %d", count)
	}
	if bytesDeleted != 128 {
		t.Fatalf("expected 128 deleted bytes, got %d", bytesDeleted)
	}
}

func TestFileService_CleanupInactiveFiles_SkipsReferenced(t *testing.T) {
	repo := &fileRepositoryMock{
		createFn:          func(ctx context.Context, file *models.File) error { return nil },
		deactivateByURLFn: func(ctx context.Context, url string) error { return nil },
		listInactiveOlderThanFn: func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
			return []models.File{
				{
					ID:         11,
					ObjectName: "used-file.png",
					URL:        "http://localhost:9000/uploads/used-file.png",
					SizeBytes:  256,
					IsActive:   false,
				},
			}, nil
		},
		isFileReferencedFn: func(ctx context.Context, file models.File) (bool, error) {
			return true, nil
		},
		deleteHardFn: func(ctx context.Context, fileID int64) error {
			t.Fatalf("delete should not be called")
			return nil
		},
	}

	storage := &objectStorageMock{
		uploadFn: func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "", nil
		},
		removeFn: func(ctx context.Context, objectName string) error {
			t.Fatalf("remove should not be called")
			return nil
		},
	}

	service := NewFileService(repo, storage)

	count, bytesDeleted, err := service.CleanupInactiveFiles(context.Background(), time.Now(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected 0 deleted files, got %d", count)
	}
	if bytesDeleted != 0 {
		t.Fatalf("expected 0 deleted bytes, got %d", bytesDeleted)
	}
}

func TestFileService_DeactivateByURL(t *testing.T) {
	called := false

	repo := &fileRepositoryMock{
		createFn: func(ctx context.Context, file *models.File) error { return nil },
		deactivateByURLFn: func(ctx context.Context, url string) error {
			called = true
			if url != "http://localhost:9000/uploads/file.png" {
				t.Fatalf("unexpected url: %s", url)
			}
			return nil
		},
		listInactiveOlderThanFn: func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
			return nil, nil
		},
		isFileReferencedFn: func(ctx context.Context, file models.File) (bool, error) { return false, nil },
		deleteHardFn:       func(ctx context.Context, fileID int64) error { return nil },
	}

	storage := &objectStorageMock{
		uploadFn: func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "", nil
		},
		removeFn: func(ctx context.Context, objectName string) error { return nil },
	}

	service := NewFileService(repo, storage)

	err := service.DeactivateByURL(context.Background(), "http://localhost:9000/uploads/file.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected deactivate to be called")
	}
}

func TestFileService_Upload_StorageError(t *testing.T) {
	repo := &fileRepositoryMock{
		createFn:          func(ctx context.Context, file *models.File) error { return nil },
		deactivateByURLFn: func(ctx context.Context, url string) error { return nil },
		listInactiveOlderThanFn: func(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
			return nil, nil
		},
		isFileReferencedFn: func(ctx context.Context, file models.File) (bool, error) { return false, nil },
		deleteHardFn:       func(ctx context.Context, fileID int64) error { return nil },
	}

	storage := &objectStorageMock{
		uploadFn: func(ctx context.Context, reader io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "", errors.New("storage failed")
		},
		removeFn: func(ctx context.Context, objectName string) error { return nil },
	}

	service := NewFileService(repo, storage)

	_, err := service.Upload(
		context.Background(),
		bytes.NewBufferString("hello-image"),
		"avatar.png",
		"image/png",
		int64(len("hello-image")),
	)
	if err == nil {
		t.Fatalf("expected error")
	}
}
