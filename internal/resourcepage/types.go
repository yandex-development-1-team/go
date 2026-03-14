package resourcepage

import (
	"encoding/json"
	"errors"
	"time"
)

// Ошибки домена (используются репозиторием и сервисом)
var (
	ErrNotFound = errors.New("resource page not found")
)

// --- API (DTO) ---

// UpdateRequest — PUT /api/v1/resources/{slug}
type UpdateRequest struct {
	Slug    string  `json:"slug,omitempty"`
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Links   *[]Link `json:"links,omitempty"`
}
type Link struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Response — GET /api/v1/resources/{slug}
type Response struct {
	Title   string  `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Links   *[]Link `json:"links,omitempty"`
}

// ResponsePublic — GET /api/v1/public/resources/{slug}
type ResponsePublic struct {
	Title   string  `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Links   *[]Link `json:"links,omitempty"`
}

// DeleteLinkRequest — DELETE /api/v1/resources/{slug}/{id}
type DeleteLinkRequest struct {
	Slug string `json:"slug"`
	ID   string `json:"id"`
}

// --- Domain (сервисный слой) ---

type ResourcePage struct {
	Slug      string          `json:"slug"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Links     []Link          `json:"-"`
	LinksJSON json.RawMessage `json:"links"`
	UpdatedAt string          `json:"updated_at"`
}

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
