package models

import (
	"errors"
	"time"
)

var (
	ErrRequestTimeout   = errors.New("request timeout")
	ErrRequestCanceled  = errors.New("request canceled")
	ErrDatabase         = errors.New("database error")
	ErrSlotOccupied     = errors.New("slot is already occupied")
	ErrInvalidInput     = errors.New("invalid input data")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrSpecProjNotFound = errors.New("special project not found")
)

type SpecialProjectDB struct {
	ID            int64     `db:"id"`
	Title         string    `db:"title"`
	Description   *string   `db:"description"` // nullable
	Image         string    `db:"image"`
	IsActiveInBot bool      `db:"is_active_in_bot"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
