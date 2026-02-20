package models

import (
	"errors"
	"time"
)

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrRequestCanceled = errors.New("request canceled")
	ErrDatabase        = errors.New("database error")
	ErrSlotOccupied    = errors.New("slot is already occupied")
	ErrInvalidInput    = errors.New("invalid input data")
	ErrBookingNotFound = errors.New("booking not found")
	ErrUserNotFound    = errors.New("user not found")
)

type User struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Grade      int       `db:"grade"`
	IsAdmin    bool      `db:"is_admin"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
type Booking struct {
	ID                int64      `db:"id"`
	UserID            int64      `db:"user_id"`
	ServiceID         int16      `db:"service_id"`
	BookingDate       time.Time  `db:"booking_date"`
	BookingTime       *time.Time `db:"booking_time"`
	GuestName         string     `db:"guest_name"`
	GuestOrganization string     `db:"guest_organization"`
	GuestPosition     string     `db:"guest_position"`
	VisitType         string     `db:"visit_type"`
	Status            string     `db:"status"`
	TrackerTicketID   string     `db:"tracker_ticket_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type UserSession struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	CurrentState string                 `json:"current_state"`
	StateData    map[string]interface{} `json:"state_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`

type BoxSolution struct {
	ID             int64
	Name           string
	Description    string
	AvailableSlots []AvailableSlot
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}
