package models

import (
	"errors"
	"time"
)

var (
	ErrBoxNotFound   = errors.New("box not found")
	ErrEventNotFound = errors.New("event not found")
)

type EventStatus string

const (
	EventStatusActive    EventStatus = "active"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

type Event struct {
	ID            int64       `json:"id" db:"id"`
	BoxID         int64       `json:"box_id" db:"box_id"`
	Date          time.Time   `json:"date" db:"event_date"`
	Time          string      `json:"time" db:"event_time"`
	TotalSlots    int         `json:"total_slots" db:"total_slots"`
	OccupiedSlots int         `json:"occupied_slots" db:"occupied_slots"`
	Status        EventStatus `json:"status" db:"status"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

type EventCreateRequest struct {
	BoxID      int64  `json:"box_id"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	TotalSlots int    `json:"total_slots"`
}

type EventListItem struct {
	Event
	BoxName string `json:"box_name" db:"box_name"`
}

type EventListResponse struct {
	Items      []EventListItem `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

type BookingShort struct {
	ID                int64  `json:"id" db:"booking_id"`
	GuestName         string `json:"guest_name" db:"guest_name"`
	GuestOrganization string `json:"guest_organization" db:"guest_organization"`
	Status            string `json:"status" db:"booking_status"`
}

type EventWithBookings struct {
	EventListItem
	Box struct {
		ID    int64  `json:"id" db:"box_id"`
		Name  string `json:"name" db:"box_name"`
		Image string `json:"image" db:"box_image"`
	} `json:"box"`
	Bookings []BookingShort `json:"bookings"`
}

type EventFilter struct {
	BoxID    *int64
	DateFrom *time.Time
	DateTo   *time.Time
	Status   *string
	Limit    int
	Offset   int
}
