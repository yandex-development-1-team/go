package models

import "time"

// SpecialProject — доменная модель спецпроекта (сервисный слой, API).
type SpecialProject struct {
	ID            int64
	Title         string
	Description   *string
	Image         string
	IsActiveInBot bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SpecialProjectDB — модель спецпроекта для БД (repository).
type SpecialProjectDB struct {
	ID            int64     `db:"id"`
	Title         string    `db:"title"`
	Description   *string   `db:"description"`
	Image         string    `db:"image"`
	IsActiveInBot bool      `db:"is_active_in_bot"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
