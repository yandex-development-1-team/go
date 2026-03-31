package dto

import (
	"time"

	"github.com/yandex-development-1-team/go/internal/models"
)

type EventCreateRequest struct {
	BoxID      int64  `json:"box_id"       binding:"required"`
	Date       string `json:"date"         binding:"required"`
	Time       string `json:"time"         binding:"required"`
	TotalSlots int    `json:"total_slots"  binding:"required,min=1"`
}

type EventListQuery struct {
	BoxID    *int64     `form:"box_id"`
	DateFrom *time.Time `form:"date_from" time_format:"2006-01-02"`
	DateTo   *time.Time `form:"date_to"   time_format:"2006-01-02"`
	Status   *string    `form:"status"`
	Limit    int        `form:"limit,default=50"`
	Offset   int        `form:"offset,default=0"`
}

type EventRow struct {
	ID            int64     `db:"id"`
	BoxID         int64     `db:"box_id"`
	Date          time.Time `db:"event_date"`
	Time          string    `db:"event_time"`
	TotalSlots    int       `db:"total_slots"`
	OccupiedSlots int       `db:"occupied_slots"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	BoxName       *string   `db:"box_name"`
}

type EventWithBookingRow struct {
	EventRow
	BookingID         *int64  `db:"booking_id"`
	GuestName         *string `db:"guest_name"`
	GuestOrganization *string `db:"guest_organization"`
	BookingStatus     *string `db:"booking_status"`
}

type EventsListResponse struct {
	Items      []models.EventListItem `json:"items"`
	Pagination models.Pagination      `json:"pagination"`
}
