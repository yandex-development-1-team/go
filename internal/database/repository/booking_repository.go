package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
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

	updateBookingStatusQuery = `
		UPDATE bookings 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`
)

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrRequestCanceled = errors.New("request canceled")
	ErrDatabase        = errors.New("database error")
	ErrSlotOccupied    = errors.New("slot is already occupied")
	ErrInvalidInput    = errors.New("invalid input data")
	ErrBookingNotFound = errors.New("booking not found")
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, b *models.Booking) (int64, error)
	GetAvailableSlots(ctx context.Context, serviceID int, date time.Time) ([]time.Time, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error
}

type BookingRepo struct {
	db *sqlx.DB
}

var _ BookingRepository = (*BookingRepo)(nil)

func NewBookingRepository(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) CreateBooking(ctx context.Context, b *models.Booking) (int64, error) {
	const operation = "create_booking"

	if err := r.validateBooking(b); err != nil {
		return 0, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, r.checkError(operation, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, advisoryLockQuery, b.ServiceID, b.BookingDate, b.BookingTime); err != nil {
		return 0, r.checkError(operation, err)
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
			return 0, ErrSlotOccupied
		}
		return 0, r.checkError(operation, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, r.checkError(operation, err)
	}

	return id, nil
}

func (r *BookingRepo) GetAvailableSlots(ctx context.Context, serviceID int, date time.Time) ([]time.Time, error) {
	var slots []time.Time
	err := r.db.SelectContext(ctx, &slots, getAvailableSlotsQuery, serviceID, date)
	if err != nil {
		return nil, r.checkError("get_available_slots", err)
	}
	return slots, nil
}

func (r *BookingRepo) GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.SelectContext(ctx, &bookings, getBookingsByUserIDQuery, userID)
	if err != nil {
		return nil, r.checkError("get_bookings_by_user", err)
	}
	return bookings, nil
}

func (r *BookingRepo) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	const operation = "update_booking_status"
	if bookingID <= 0 || status == "" {
		return ErrInvalidInput
	}

	res, err := r.db.ExecContext(ctx, updateBookingStatusQuery, status, bookingID)
	if err != nil {
		return r.checkError(operation, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return r.checkError(operation, err)
	}
	if affected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepo) validateBooking(b *models.Booking) error {
	if b == nil || b.UserID <= 0 || b.GuestName == "" || b.ServiceID <= 0 || b.BookingDate.IsZero() {
		return ErrInvalidInput
	}
	return nil
}

func (r *BookingRepo) checkError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		logger.Error("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return ErrRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Error("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return ErrRequestTimeout
	}

	logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return ErrDatabase
}
