package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           int64     `db:"id"`
	UUID         uuid.UUID `db:"uuid"`
	ObjectName   string    `db:"object_name"`
	OriginalName string    `db:"original_name"`
	URL          string    `db:"url"`
	MimeType     string    `db:"mime_type"`
	SizeBytes    int64     `db:"size_bytes"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
