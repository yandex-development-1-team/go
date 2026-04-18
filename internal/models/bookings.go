package models

import "time"

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

type BookingAPI struct {
	ID                int64     `db:"id"`
	UserID            int64     `db:"user_id"`
	ServiceID         int16     `db:"service_id"`
	ServiceName       string    `db:"service_name"`
	BookingDate       string    `db:"booking_date"`
	BookingTime       string    `db:"booking_time"`
	GuestName         string    `db:"guest_name"`
	GuestOrganization string    `db:"guest_organization"`
	GuestContact      string    `db:"guest_contact"`
	GuestPosition     string    `db:"guest_position"`
	Status            string    `db:"status"`
	ManagerID         int64     `db:"manager_id"`
	ManagerName       string    `db:"manager_name"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type BookingList struct {
	Items  []BookingAPI
	Total  int
	Limit  int
	Offset int
}
