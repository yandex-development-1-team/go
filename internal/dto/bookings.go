package dto

import "time"

type BookingsID struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type BookingAPIRaw struct {
	ID                int64      `db:"id"`
	UserID            int64      `db:"user_id"`
	ServiceID         int16      `db:"service_id"`
	ServiceName       string     `db:"service_name"`
	BookingDate       time.Time  `db:"booking_date"`
	BookingTime       *time.Time `db:"booking_time"`
	GuestName         string     `db:"guest_name"`
	GuestOrganization string     `db:"guest_organization"`
	GuestContact      string     `db:"guest_contact"`
	GuestPosition     string     `db:"guest_position"`
	Status            string     `db:"status"`
	ManagerID         int64      `db:"manager_id"`
	ManagerName       string     `db:"manager_name"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type BookingDetailResponse struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"user_id"`
	ServiceID         int16     `json:"service_id"`
	ServiceName       string    `json:"service_name"`
	BookingDate       string    `json:"booking_date"`
	BookingTime       string    `json:"booking_time"`
	GuestName         string    `json:"guest_name"`
	GuestOrganization string    `json:"guest_organization"`
	GuestContact      string    `json:"guest_contact"`
	GuestPosition     string    `json:"guest_position"`
	Status            string    `json:"status"`
	ManagerID         int64     `json:"manager_id"`
	ManagerName       string    `json:"manager_name"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type BookingListItem struct {
	ID           int64     `db:"id" json:"id"`
	Status       string    `db:"status" json:"status"`
	CustomerName string    `db:"guest_name" json:"guest_name"`
	ContactInfo  string    `db:"guest_contact" json:"guest_contact"`
	ServiceName  string    `db:"service_name" json:"service_name"`
	ManagerID    int64     `db:"manager_id" json:"manager_id"`
	ManagerName  string    `db:"manager_name" json:"manager_name"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type BookingListResponse struct {
	Items      []BookingListItem `json:"items"`
	Pagination Pagination        `json:"pagination"`
}

type BookingRow struct {
	ID           int64     `db:"id"`
	Status       string    `db:"status"`
	CustomerName string    `db:"guest_name"`
	ContactInfo  string    `db:"guest_contact"`
	ServiceName  string    `db:"service_name"`
	ManagerID    int64     `db:"manager_id"`
	ManagerName  string    `db:"manager_name"`
	CreatedAt    time.Time `db:"created_at"`
	Total        int       `db:"total"`
}
