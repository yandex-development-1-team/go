package models

import (
	"errors"
)

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
