package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const (
	advisoryLockQuery = `SELECT pg_advisory_xact_lock(hashtext(concat($1::text, $2::text, $3::text)))`

	createBookingAtomicQuery = `
		INSERT INTO bookings (
			user_id, service_id, booking_date, booking_time, 
			guest_name, guest_organization, guest_position, 
			visit_type, tracker_ticket_id
		) 
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		WHERE NOT EXISTS (
			SELECT 1 FROM bookings 
			WHERE service_id = $2 
			  AND booking_date = $3 
			  AND booking_time IS NOT DISTINCT FROM $4 
			  AND status = 'confirmed'
		)
		RETURNING id`

	getAvailableSlotsQuery = `
		SELECT booking_time 
		FROM bookings 
		WHERE service_id = $1 
		  AND booking_date = $2 
		  AND status != 'confirmed'
		ORDER BY booking_time ASC`

	getBookingsByUserIDQuery = `
		SELECT id, user_id, service_id, booking_date, booking_time, 
		       guest_name, guest_organization, guest_position, 
		       visit_type, status, tracker_ticket_id, created_at, updated_at
		FROM bookings 
		WHERE user_id = $1 
		ORDER BY booking_date DESC, booking_time DESC`

	getBookingsByUserIDExtQuery = `
		SELECT b.id, b.user_id, b.service_id, s.name as service_name,
		       b.booking_date, b.booking_time, 
		       b.guest_name, b.guest_organization, b.guest_position, 
		       b.visit_type, b.status, b.tracker_ticket_id, 
		       b.created_at, b.updated_at
		FROM bookings b
		LEFT JOIN services s ON b.service_id = s.id
		WHERE b.user_id = $1 
		ORDER BY b.booking_date ASC, b.booking_time ASC`

	updateBookingStatusQuery = `
		UPDATE bookings 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`
)

type BookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) CreateBooking(ctx context.Context, b *models.Booking) (int64, error) {
	const operation = "create_booking"

	return repository.WithDBMetricsValue(operation, func() (int64, error) {

		if err := r.validateBooking(b); err != nil {
			return 0, err
		}

		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return 0, err
		}
		defer func() { _ = tx.Rollback() }()

		if _, err := tx.ExecContext(ctx, advisoryLockQuery, b.ServiceID, b.BookingDate, b.BookingTime); err != nil {
			return 0, err
		}

		var id int64
		err = tx.QueryRowContext(ctx, createBookingAtomicQuery,
			b.UserID,
			b.ServiceID,
			b.BookingDate,
			b.BookingTime,
			b.GuestName,
			b.GuestOrganization,
			b.GuestPosition,
			b.VisitType,
			b.TrackerTicketID,
		).Scan(&id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, models.ErrSlotOccupied
			}
			return 0, err
		}

		if err = tx.Commit(); err != nil {
			return 0, err
		}

		return id, nil
	})
}

func (r *BookingRepo) GetAvailableSlots(ctx context.Context, serviceID int, date time.Time) ([]time.Time, error) {
	const operation = "get_available_slots"
	var slots []time.Time

	return repository.WithDBMetricsValue(operation, func() ([]time.Time, error) {

		err := r.db.SelectContext(ctx, &slots, getAvailableSlotsQuery, serviceID, date)
		if err != nil {
			return nil, err
		}
		return slots, nil
	})
}

func (r *BookingRepo) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	const operation = "get_booking_by_id"
	var bookings []models.Booking

	return repository.WithDBMetricsValue(operation, func() ([]models.Booking, error) {

		err := r.db.SelectContext(ctx, &bookings, getBookingsByUserIDQuery, userID)
		if err != nil {
			return nil, err
		}
		return bookings, nil
	})
}

func (r *BookingRepo) GetExtBookingsByUserID(ctx context.Context, userID int64) ([]models.BookingExt, error) {
	const operation = "get_booking_by_id"
	var bookings []models.BookingExt

	return repository.WithDBMetricsValue(operation, func() ([]models.BookingExt, error) {

		err := r.db.SelectContext(ctx, &bookings, getBookingsByUserIDExtQuery, userID)
		if err != nil {
			return nil, err
		}
		return bookings, nil
	})
}

func (r *BookingRepo) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	const operation = "update_booking_status"

	return repository.WithDBMetrics(operation, func() error {

		if bookingID <= 0 || status == "" {
			return models.ErrInvalidInput
		}

		res, err := r.db.ExecContext(ctx, updateBookingStatusQuery, status, bookingID)
		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return models.ErrBookingNotFound
		}
		return nil
	})
}

func (r *BookingRepo) validateBooking(b *models.Booking) error {
	if b == nil || b.UserID <= 0 || b.GuestName == "" || b.ServiceID <= 0 || b.BookingDate.IsZero() {
		return models.ErrInvalidInput
	}
	return nil
}
