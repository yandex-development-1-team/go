package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

// FileService provides file upload, deactivation, and cleanup operations.
type FileService struct {
	repo    FileRepository
	storage ObjectStorage
}

// NewFileService creates a new FileService.
func NewFileService(repo FileRepository, storage ObjectStorage) *FileService {
	return &FileService{repo: repo, storage: storage}
}

// Upload uploads a file to object storage and saves its metadata in the files table.
func (s *FileService) Upload(
	ctx context.Context,
	reader io.Reader,
	originalName string,
	contentType string,
	size int64,
) (*dto.FileUploadResponse, error) {
	if reader == nil {
		return nil, fmt.Errorf("file reader is nil")
	}
	originalName = strings.TrimSpace(originalName)
	if originalName == "" {
		return nil, fmt.Errorf("original filename is empty")
	}

	fileUUID := uuid.New()
	ext := strings.ToLower(filepath.Ext(originalName))
	objectName := fileUUID.String()
	if ext != "" {
		objectName += ext
	}

	var buf bytes.Buffer
	written, err := io.Copy(&buf, reader)
	if err != nil {
		return nil, fmt.Errorf("read upload payload: %w", err)
	}

	if written == 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}

	if size <= 0 {
		size = written
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileURL, err := s.storage.UploadFile(ctx, bytes.NewReader(buf.Bytes()), objectName, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload file to storage: %w", err)
	}

	now := time.Now().UTC()
	file := &models.File{
		UUID:         fileUUID,
		ObjectName:   objectName,
		OriginalName: originalName,
		URL:          fileURL,
		MimeType:     contentType,
		SizeBytes:    size,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("save file metadata: %w", err)
	}
	return &dto.FileUploadResponse{
		UUID:     file.UUID.String(),
		URL:      file.URL,
		Filename: file.OriginalName,
	}, nil
}

// DeactivateByURL marks a file as inactive by its public URL.
func (s *FileService) DeactivateByURL(ctx context.Context, fileURL string) error {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return nil
	}

	if err := s.repo.DeactivateByURL(ctx, fileURL); err != nil {
		return fmt.Errorf("deactivate file by URL: %w", err)
	}
	return nil
}

// CleanupInactiveFiles removes inactive orphaned files from storage and database.
func (s *FileService) CleanupInactiveFiles(ctx context.Context, olderThan time.Time, limit int) (int, int64, error) {
	if limit <= 0 {
		limit = 100
	}

	files, err := s.repo.ListInactiveOlderThan(ctx, olderThan, limit)
	if err != nil {
		return 0, 0, fmt.Errorf("list inactive files: %w", err)
	}

	var deletedCount int
	var deletedBytes int64

	for _, file := range files {
		referenced, err := s.repo.IsFileReferenced(ctx, file)
		if err != nil {
			return deletedCount, deletedBytes, fmt.Errorf("check file referenced: %w", err)
		}
		if referenced {
			continue
		}
		if err := s.storage.RemoveFile(ctx, file.ObjectName); err != nil {
			return deletedCount, deletedBytes, fmt.Errorf("remove file from storage: %w", err)
		}

		if err := s.repo.DeleteHard(ctx, file.ID); err != nil {
			return deletedCount, deletedBytes, fmt.Errorf("delete file from database: %w", err)
		}

		deletedCount++
		deletedBytes += file.SizeBytes
	}
	return deletedCount, deletedBytes, nil
}
