package models

import (
	"encoding/json"
	"errors"
	"time"
)

type ResourcePageDB struct {
	Slug      string          `db:"slug"`
	Title     string          `db:"title"`
	Content   *string         `db:"content"`
	LinksJSON json.RawMessage `db:"links"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrRequestCanceled = errors.New("request canceled")
	ErrDatabase        = errors.New("database error")
	ErrSlotOccupied    = errors.New("slot is already occupied")
	ErrInvalidInput    = errors.New("invalid input data")
	ErrBookingNotFound = errors.New("booking not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrResPageNotFound = errors.New("resource page not found")
)
