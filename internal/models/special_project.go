package models

import (
	"errors"
	"time"
)

var ErrSpecialProjectAlreadyExists = errors.New("special project with such title already exists")

type SpecialProject struct {
	ID          int64         `json:"id,omitempty"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	Image       string        `json:"image,omitempty"`
	Status      ServiceStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty"`
}

type SpecialProjectDB struct {
	ID          int64         `db:"id"`
	Title       string        `db:"title"`
	Description *string       `db:"description"`
	Image       string        `db:"image"`
	Status      ServiceStatus `db:"status"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
}

type SpecialProjectUpdate struct {
	Title       string        `db:"title"`
	Description *string       `db:"description"`
	Image       string        `db:"image"`
	Status      ServiceStatus `db:"status"`
}
