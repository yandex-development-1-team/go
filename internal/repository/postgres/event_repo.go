package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	createEventQuery = `
		INSERT INTO events (box_id, event_date, event_time, total_slots, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, box_id, event_date, event_time, total_slots, occupied_slots, status, created_at, updated_at`

	getEventByIDQuery = `
		SELECT 
			e.*, b.title as box_name,
			bk.id as booking_id, bk.guest_name, bk.guest_organization, bk.status as booking_status
		FROM events e
		JOIN boxes b ON e.box_id = b.id
		LEFT JOIN bookings bk ON e.id = bk.event_id
		WHERE e.id = $1`
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
	var row dto.EventRow
	err := r.db.GetContext(ctx, &row, createEventQuery,
		event.BoxID, event.Date, event.Time, event.TotalSlots, event.Status)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return nil, models.ErrBoxNotFound
		}
		return nil, fmt.Errorf("repo create event: %w", err)
	}

	return toEventDomain(&row), nil
}

func (r *eventRepo) GetByID(ctx context.Context, id int64) (*models.EventWithBookings, error) {
	rows, err := r.db.QueryxContext(ctx, getEventByIDQuery, id)
	if err != nil {
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	defer rows.Close()

	var result *models.EventWithBookings
	for rows.Next() {
		var row dto.EventWithBookingRow
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("struct scan: %w", err)
		}

		if result == nil {
			result = &models.EventWithBookings{
				EventListItem: models.EventListItem{
					Event:   *toEventDomain(&row.EventRow),
					BoxName: derefString(row.BoxName),
				},
				Bookings: []models.BookingShort{},
			}
			result.Box.ID = row.BoxID
			result.Box.Name = derefString(row.BoxName)
		}

		if row.BookingID != nil {
			result.Bookings = append(result.Bookings, models.BookingShort{
				ID:                *row.BookingID,
				GuestName:         derefString(row.GuestName),
				GuestOrganization: derefString(row.GuestOrganization),
				Status:            derefString(row.BookingStatus),
			})
		}
	}

	if result == nil {
		return nil, models.ErrEventNotFound
	}

	return result, nil
}

func (r *eventRepo) List(ctx context.Context, f models.EventFilter) ([]models.EventListItem, int, error) {
	args := map[string]interface{}{
		"limit":  f.Limit,
		"offset": f.Offset,
	}
	where := "1=1"

	if f.BoxID != nil {
		where += " AND e.box_id = :box_id"
		args["box_id"] = *f.BoxID
	}

	if f.DateFrom != nil {
		where += " AND e.event_date >= :date_from"
		args["date_from"] = f.DateFrom
	}
	if f.DateTo != nil {
		where += " AND e.event_date <= :date_to"
		args["date_to"] = f.DateTo
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events e WHERE %s", where)

	nstmtCount, err := r.db.PrepareNamedContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer nstmtCount.Close()

	if err := nstmtCount.GetContext(ctx, &total, args); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
        SELECT e.*, b.title as box_name 
        FROM events e 
        JOIN boxes b ON e.box_id = b.id 
        WHERE %s 
        ORDER BY e.event_date DESC, e.event_time DESC 
        LIMIT :limit OFFSET :offset`, where)

	nstmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer nstmt.Close()

	var rows []dto.EventRow
	if err := nstmt.SelectContext(ctx, &rows, args); err != nil {
		return nil, 0, err
	}

	items := make([]models.EventListItem, len(rows))
	for i, row := range rows {
		items[i] = models.EventListItem{
			Event:   *toEventDomain(&row),
			BoxName: derefString(row.BoxName),
		}
	}

	return items, total, nil
}
func toEventDomain(row *dto.EventRow) *models.Event {
	return &models.Event{
		ID:            row.ID,
		BoxID:         row.BoxID,
		Date:          row.Date,
		Time:          row.Time,
		TotalSlots:    row.TotalSlots,
		OccupiedSlots: row.OccupiedSlots,
		Status:        models.EventStatus(row.Status),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
