package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

// FileRepository provides PostgreSQL access for file metadata.
type FileRepository struct {
	db *sqlx.DB
}

// NewFileRepository creates a new FileRepository.
func NewFileRepository(db *sqlx.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	const query = `
		INSERT INTO files (
			uuid,
			object_name,
			original_name,
			url,
			mime_type,
			size_bytes,
			is_active,
			created_at,
			updated_at
		) VALUES (
			:uuid,
			:object_name,
			:original_name,
			:url,
			:mime_type,
			:size_bytes,
			:is_active,
			:created_at,
			:updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, file)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	return nil
}

func (r *FileRepository) GetByUUID(ctx context.Context, fileUUID uuid.UUID) (*models.File, error) {
	const query = `
		SELECT
			id,
			uuid,
			object_name,
			original_name,
			url,
			mime_type,
			size_bytes,
			is_active,
			created_at,
			updated_at
		FROM files
		WHERE uuid = $1
		LIMIT 1
	`

	var file models.File
	err := r.db.GetContext(ctx, &file, query, fileUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get file by uuid: %w", err)
	}
	return &file, nil
}

func (r *FileRepository) GetByURL(ctx context.Context, url string) (*models.File, error) {
	const query = `
		SELECT
			id,
			uuid,
			object_name,
			original_name,
			url,
			mime_type,
			size_bytes,
			is_active,
			created_at,
			updated_at
		FROM files
		WHERE url = $1
		LIMIT 1
	`

	var file models.File
	err := r.db.GetContext(ctx, &file, query, url)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get file by url: %w", err)
	}
	return &file, nil
}

func (r *FileRepository) DeactivateByURL(ctx context.Context, url string) error {
	const query = `
		UPDATE files
		SET
			is_active = false,
			updated_at = NOW()
		WHERE url = $1
		  AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, url)
	if err != nil {
		return fmt.Errorf("deactivate file by url: %w", err)
	}
	return nil
}

func (r *FileRepository) ListInactiveOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]models.File, error) {
	const query = `
		SELECT
			id,
			uuid,
			object_name,
			original_name,
			url,
			mime_type,
			size_bytes,
			is_active,
			created_at,
			updated_at
		FROM files
		WHERE is_active = false
		  AND updated_at < $1
		ORDER BY updated_at ASC
		LIMIT $2
	`

	var files []models.File
	err := r.db.SelectContext(ctx, &files, query, olderThan, limit)
	if err != nil {
		return nil, fmt.Errorf("list inactive files: %w", err)
	}
	return files, nil
}

func (r *FileRepository) IsFileReferenced(ctx context.Context, file models.File) (bool, error) {
	queries := []string{
		`SELECT EXISTS(SELECT 1 FROM special_projects WHERE image = $1)`,
		`SELECT EXISTS(SELECT 1 FROM resource_pages WHERE links::text LIKE '%' || $1 || '%')`,
	}
	for _, query := range queries {
		var exists bool
		err := r.db.GetContext(ctx, &exists, query, file.URL)
		if err != nil {
			return false, fmt.Errorf("check file reference: %w", err)
		}
		if exists {
			return true, nil
		}
	}

	const servicesColumnQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'services'
			  AND column_name = 'image'
		)
	`

	var hasServicesImage bool
	if err := r.db.GetContext(ctx, &hasServicesImage, servicesColumnQuery); err != nil {
		return false, fmt.Errorf("check services.image existence: %w", err)
	}
	if hasServicesImage {
		var exists bool
		err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM services WHERE image = $1)`, file.URL)
		if err != nil {
			return false, fmt.Errorf("check services image reference: %w", err)
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func (r *FileRepository) DeleteHard(ctx context.Context, fileID int64) error {
	const query = `DELETE FROM files WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("delete hard file: %w", err)
	}

	return nil
}
