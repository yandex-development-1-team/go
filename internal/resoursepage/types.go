package resoursepage

import (
	"encoding/json"
	"time"
)

// Ошибки домена (используются репозиторием и сервисом)
var ()

// --- API (DTO) ---

// CreateRequest — тело POST /special-projects

// Response — ответ GET /special-projects/{id}

// --- Domain (сервисный слой) ---

// --- Repository (БД) ---

// DB — модель для чтения/записи в БД
type DB struct {
	Slug      string          `db:"slug"`
	Title     string          `db:"title"`
	Content   *string         `db:"content"`
	LinksJSON json.RawMessage `db:"links"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}
