package models

import (
	"errors"
	"time"
)

var ErrSpecialProjectAlreadyExists = errors.New("special project with such title already exists")

type SpecialProject struct {
	ID          int64         `json:"id,omitempty"`
	Title       string        `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string       `json:"description,omitempty" binding:"omitempty,max=1000"`
	Image       string        `json:"image,omitempty" binding:"omitempty,httpurl,max=500"`
	Status      ServiceStatus `json:"status" binding:"omitempty,oneof=active inactive"`
	CreatedAt   time.Time     `json:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty"`
}

type SpecialProjectDB struct {
	ID          int64         `db:"id"`
	Title       string        `db:"title"`
	Description *string       `db:"description"`
	Image       *string       `db:"image"`
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
