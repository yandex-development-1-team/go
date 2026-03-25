package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const analyticsExportLimit = 10_000

const (
	getBoxesAnalyticsQuery = `
		SELECT
			s.id           AS service_id,
			s.name         AS service_name,
			COUNT(b.id)    AS total_bookings,
			COUNT(b.id) FILTER (WHERE b.status = 'confirmed')  AS confirmed_bookings,
			COUNT(b.id) FILTER (WHERE b.status = 'cancelled')  AS cancelled_bookings,
			CASE WHEN COUNT(b.id) > 0
				THEN ROUND(
					COUNT(b.id) FILTER (WHERE b.status = 'cancelled')::numeric
					/ COUNT(b.id) * 100, 2
				)
				ELSE 0
			END             AS cancellation_rate
		FROM services s
		LEFT JOIN bookings b
			ON  b.service_id = s.id
			AND ($1::date IS NULL OR b.booking_date >= $1::date)
			AND ($2::date IS NULL OR b.booking_date <= $2::date)
		WHERE s.box_solution = TRUE
		GROUP BY s.id, s.name
		ORDER BY s.name
		LIMIT $3`

	getUsersAnalyticsQuery = `
		SELECT
			u.id            AS user_id,
			u.first_name,
			u.last_name,
			u.email,
			COUNT(b.id)     AS total_bookings,
			u.created_at    AS registered_at
		FROM users u
		LEFT JOIN bookings b
			ON  b.user_id = u.id
			AND ($1::date IS NULL OR b.booking_date >= $1::date)
			AND ($2::date IS NULL OR b.booking_date <= $2::date)
		GROUP BY u.id, u.first_name, u.last_name, u.email, u.created_at
		ORDER BY u.created_at DESC
		LIMIT $3`
)

type AnalyticsRepo struct {
	db *sqlx.DB
}

func NewAnalyticsRepo(db *sqlx.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) GetBoxesAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsBoxRow, error) {
	const operation = "get_boxes_analytics"
	var rows []dto.AnalyticsBoxRow
	return repository.WithDBMetricsValue(operation, func() ([]dto.AnalyticsBoxRow, error) {
		err := r.db.SelectContext(ctx, &rows, getBoxesAnalyticsQuery, dateFrom, dateTo, analyticsExportLimit)
		return rows, err
	})
}

func (r *AnalyticsRepo) GetUsersAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsUserRow, error) {
	const operation = "get_users_analytics"
	var rows []dto.AnalyticsUserRow
	return repository.WithDBMetricsValue(operation, func() ([]dto.AnalyticsUserRow, error) {
		err := r.db.SelectContext(ctx, &rows, getUsersAnalyticsQuery, dateFrom, dateTo, analyticsExportLimit)
		return rows, err
	})
}
