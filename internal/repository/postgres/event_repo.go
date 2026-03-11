package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/models"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) (*models.Event, error)
	GetByID(ctx context.Context, id int64) (*models.EventWithBookings, error)
	List(ctx context.Context, filter models.EventFilter) ([]models.EventListItem, int, error)
}

type eventRepo struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) EventRepository {
	return &eventRepo{db: db}
}

func (r *eventRepo) Create(ctx context.Context, event *models.Event) (*models.Event, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM boxes WHERE id = $1)`
	if err := r.db.GetContext(ctx, &exists, checkQuery, event.BoxID); err != nil {
		return nil, fmt.Errorf("check box exists: %w", err)
	}
	if !exists {
		return nil, models.ErrBoxNotFound
	}

	query := `
        INSERT INTO events (box_id, event_date, event_time, total_slots, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, box_id, event_date, event_time, total_slots, occupied_slots, status, created_at, updated_at
    `

	err := r.db.QueryRowxContext(ctx, query,
		event.BoxID, event.Date, event.Time, event.TotalSlots, event.Status,
	).StructScan(event)

	if err != nil {
		return nil, fmt.Errorf("repo create event: %w", err)
	}

	return event, nil
}

func (r *eventRepo) List(ctx context.Context, f models.EventFilter) ([]models.EventListItem, int, error) {
	args := make(map[string]interface{})
	where := "1=1"

	if f.BoxID != nil {
		where += " AND e.box_id = :box_id"
		args["box_id"] = *f.BoxID
	}
	if f.DateFrom != nil {
		where += " AND e.event_date >= :date_from"
		args["date_from"] = *f.DateFrom
	}
	if f.DateTo != nil {
		where += " AND e.event_date <= :date_to"
		args["date_to"] = *f.DateTo
	}
	if f.Status != nil {
		where += " AND e.status = :status"
		args["status"] = *f.Status
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events e WHERE %s", where)

	nstmtCount, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("prepare count query: %w", err)
	}
	if err := nstmtCount.GetContext(ctx, &total, args); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	query := fmt.Sprintf(`
        SELECT e.*, b.title as box_name 
        FROM events e
        JOIN boxes b ON e.box_id = b.id
        WHERE %s
        ORDER BY e.event_date DESC, e.event_time DESC
        LIMIT :limit OFFSET :offset`, where)

	args["limit"] = f.Limit
	args["offset"] = f.Offset

	var items []models.EventListItem
	nstmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	if err := nstmt.SelectContext(ctx, &items, args); err != nil {
		return nil, 0, fmt.Errorf("select events: %w", err)
	}

	return items, total, nil
}

func (r *eventRepo) GetByID(ctx context.Context, id int64) (*models.EventWithBookings, error) {
	query := `
        SELECT 
            e.*, 
            b.title as box_name,
            bk.id as booking_id, bk.guest_name, bk.guest_organization, bk.status as booking_status
        FROM events e
        JOIN boxes b ON e.box_id = b.id
        LEFT JOIN bookings bk ON e.id = bk.event_id
        WHERE e.id = $1
    `

	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result *models.EventWithBookings
	for rows.Next() {
		if result == nil {
			result = &models.EventWithBookings{}
			result.Bookings = []models.BookingShort{}
		}

		type rowStruct struct {
			models.EventListItem
			BookingID     sql.NullInt64  `db:"booking_id"`
			GuestName     sql.NullString `db:"guest_name"`
			GuestOrg      sql.NullString `db:"guest_organization"`
			BookingStatus sql.NullString `db:"booking_status"`
		}

		var row rowStruct
		if err := rows.StructScan(&row); err != nil {
			return nil, err
		}

		if result.ID == 0 {
			result.EventListItem = row.EventListItem
			result.Box.ID = row.BoxID
			result.Box.Name = row.BoxName
		}

		if row.BookingID.Valid {
			result.Bookings = append(result.Bookings, models.BookingShort{
				ID:                row.BookingID.Int64,
				GuestName:         row.GuestName.String,
				GuestOrganization: row.GuestOrg.String,
				Status:            row.BookingStatus.String,
			})
		}
	}

	if result == nil {
		return nil, models.ErrEventNotFound
	}

	return result, nil
}
