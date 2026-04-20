package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const (
	createBookingAtomicQuery = `
	INSERT INTO bookings (
    user_id, service_id, booking_date, booking_time, 
    guest_name, guest_organization, guest_position, 
    visit_type, tracker_ticket_id, manager_id
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
    (
        SELECT s.id
        FROM staff s
        LEFT JOIN applications a ON a.manager_id = s.id
            AND a.status != 'cancelled'
        LEFT JOIN bookings b ON b.manager_id = s.id
            AND b.status != 'cancelled'
        WHERE s.role IN ('manager_1', 'manager_2', 'manager_3', 'admin')
        GROUP BY s.id
        ORDER BY COUNT(a.id) + COUNT(b.id) ASC
        LIMIT 1
    )
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
	SET status = $1, updated_at = NOW()
	WHERE id = $2 AND deleted_at IS NULL
			`

	getBookingById = `
		SELECT b.id, b.user_id, b.service_id, b.booking_date, b.booking_time,
			b.guest_name, b.guest_organization, b.guest_position,
			b.status, b.created_at, b.updated_at,
			COALESCE(b.manager_id, 0) AS manager_id,
			COALESCE(s.first_name || ' ' || s.last_name, '') AS manager_name,
			u.username AS guest_contact,
			sv.name AS service_name
		FROM bookings b
		LEFT JOIN staff s ON s.id = b.manager_id
		LEFT JOIN users u ON b.user_id = u.telegram_id
		LEFT JOIN services sv ON sv.id = b.service_id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`
	listBookingsBaseQuery = `
		SELECT 
				b.id, b.status, b.guest_name, b.created_at,
				COALESCE(b.manager_id, 0) AS manager_id,
				COALESCE(s.first_name || ' ' || s.last_name, '') AS manager_name,
        sv.name AS service_name, u.username AS guest_contact,
				COUNT(*) OVER() AS total
		FROM bookings b
		LEFT JOIN staff s ON s.id = b.manager_id
		LEFT JOIN services sv ON sv.id = b.service_id
    LEFT JOIN users u ON b.user_id = u.telegram_id
		`

	deleteBookings = `
	UPDATE bookings 
	SET deleted_at = NOW(), updated_at = NOW() 
	WHERE id = $1`
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

func (r *BookingRepo) GetBookingById(ctx context.Context, id int64) (*models.BookingAPI, error) {
	var booking dto.BookingAPIRaw
	err := sqlx.GetContext(ctx, r.getDB(ctx), &booking, getBookingById, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrBookingNotFound
		}
		return nil, err
	}

	return toBookingDomainModel(&booking), nil
}

func (r *BookingRepo) GetBookingsList(ctx context.Context, filter *models.ApplicationFilter) (*models.BookingList, error) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 5)
	i := 1

	conditions = append(conditions, "b.deleted_at IS NULL")

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("b.status = $%d", i))
		args = append(args, filter.Status)
		i++
	}
	if filter.ManagerID != 0 {
		conditions = append(conditions, fmt.Sprintf("b.manager_id = $%d", i))
		args = append(args, filter.ManagerID)
		i++
	}
	if filter.CustomerName != "" {
		conditions = append(conditions, fmt.Sprintf("b.guest_name ILIKE $%d", i))
		args = append(args, "%"+filter.CustomerName+"%")
		i++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, filter.Limit, filter.Offset)
	dataSQL := listBookingsBaseQuery + where +
		fmt.Sprintf(" ORDER BY b.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)

	var rows []dto.BookingRow
	if err := r.db.SelectContext(ctx, &rows, dataSQL, args...); err != nil {
		return nil, err
	}

	apps := make([]models.BookingAPI, len(rows))
	for i, row := range rows {
		apps[i] = models.BookingAPI{
			ID:           row.ID,
			Status:       row.Status,
			ManagerID:    row.ManagerID,
			ManagerName:  row.ManagerName,
			GuestName:    row.CustomerName,
			ServiceName:  row.ServiceName,
			GuestContact: row.ContactInfo,
			CreatedAt:    row.CreatedAt,
		}
	}

	total := 0
	if len(rows) > 0 {
		total = rows[0].Total
	}

	return &models.BookingList{
		Items:  apps,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (r *BookingRepo) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	result, err := r.getDB(ctx).ExecContext(ctx, updateBookingStatusQuery, status, id)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return models.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepo) DeleteBooking(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, deleteBookings, id)
	if err != nil {
		return fmt.Errorf("delete booking: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return models.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepo) validateBooking(b *models.Booking) error {
	if b == nil || b.UserID <= 0 || b.GuestName == "" || b.ServiceID <= 0 || b.BookingDate.IsZero() {
		return models.ErrInvalidInput
	}
	return nil
}

func toBookingDomainModel(b *dto.BookingAPIRaw) *models.BookingAPI {
	bookingTime := ""
	if b.BookingTime != nil {
		bookingTime = b.BookingTime.Format("15:04")
	}

	return &models.BookingAPI{
		ID:                b.ID,
		UserID:            b.UserID,
		ServiceID:         b.ServiceID,
		ServiceName:       b.ServiceName,
		BookingDate:       b.BookingDate.Format("2006-01-02"),
		BookingTime:       bookingTime,
		GuestName:         b.GuestName,
		GuestOrganization: b.GuestOrganization,
		GuestContact:      "@" + b.GuestContact,
		GuestPosition:     b.GuestPosition,
		Status:            b.Status,
		ManagerID:         b.ManagerID,
		ManagerName:       b.ManagerName,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}

func (r *BookingRepo) getDB(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctxutil.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
