package specialproject

import (
	"errors"
	"time"
)

// Ошибки домена (используются репозиторием и сервисом)
var (
	ErrNotFound      = errors.New("special project not found")
	ErrAlreadyExists = errors.New("special project with such title already exists")
)

// --- API (DTO) ---

// CreateRequest — тело POST /special-projects
type CreateRequest struct {
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Image         string  `json:"image,omitempty"`
	IsActiveInBot bool    `json:"is_active_in_bot"`
}

// ListItem — элемент списка GET /special-projects
type ListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	IsActiveInBot bool   `json:"is_active_in_bot"`
}

// Response — ответ GET /special-projects/{id}
type Response struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Image         string    `json:"image,omitempty"`
	IsActiveInBot bool      `json:"is_active_in_bot"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// --- Domain (сервисный слой) ---

// Project — доменная сущность (json-теги для PUT /special-projects/{id})
type Project struct {
	ID            int64     `json:"id,omitempty"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Image         string    `json:"image,omitempty"`
	IsActiveInBot bool      `json:"is_active_in_bot"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

// --- Repository (БД) ---

// DB — модель для чтения/записи в БД
type DB struct {
	ID            int64     `db:"id"`
	Title         string    `db:"title"`
	Description   *string   `db:"description"`
	Image         string    `db:"image"`
	IsActiveInBot bool      `db:"is_active_in_bot"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Update — поля для UPDATE (без ID и timestamps)
type Update struct {
	Title         string  `db:"title"`
	Description   *string `db:"description"`
	Image         string  `db:"image"`
	IsActiveInBot bool    `db:"is_active_in_bot"`
}
