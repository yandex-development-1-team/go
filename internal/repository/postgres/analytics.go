package postgres

import (
	"context"
	"encoding/json"
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

	getAnalyticsOverviewQuery = `
    SELECT 
        COUNT(DISTINCT b.service_id) as total_events,
		COUNT(b.id) * 1000 as revenue,
        COUNT(b.id) as total_bookings,
        COUNT(DISTINCT b.user_id) as total_users,
        CASE 
            WHEN COUNT(b.id) > 0 
            THEN ROUND(COUNT(CASE WHEN b.status = 'confirmed' THEN 1 END)::numeric / COUNT(b.id) * 100, 2)
            ELSE 0 
        END as attendance_rate,
        CASE 
            WHEN COUNT(DISTINCT b.service_id) > 0 
            THEN ROUND(COUNT(b.id)::numeric / COUNT(DISTINCT b.service_id), 2)
            ELSE 0 
        END as average_attendance
    FROM bookings b
    WHERE ($1::date IS NULL OR b.booking_date >= $1::date)
      AND ($2::date IS NULL OR b.booking_date <= $2::date)
`

	getBoxesAnalyticsExtendedQuery = `
    SELECT
        s.id AS box_id,
        s.name AS box_name,
        COUNT(DISTINCT b.service_id) AS total_events,
		COUNT(CASE WHEN b.status = 'confirmed' THEN 1 END) as confirmed_bookings,
		COUNT(CASE WHEN b.status = 'cancelled' THEN 1 END) as cancelled_bookings,
        COUNT(b.id) AS total_bookings,
        CASE 
            WHEN COUNT(b.id) > 0 
            THEN ROUND(COUNT(CASE WHEN b.status = 'confirmed' THEN 1 END)::numeric / COUNT(b.id) * 100, 2)
            ELSE 0 
        END AS attendance_rate,
        CASE 
            WHEN COUNT(DISTINCT b.service_id) > 0 
            THEN ROUND(COUNT(b.id)::numeric / COUNT(DISTINCT b.service_id), 2)
            ELSE 0 
        END AS average_attendance,
        COUNT(*) FILTER (WHERE b.status = 'confirmed') * 1000 as revenue,
        CASE 
            WHEN COUNT(b.id) > 0 
            THEN ROUND(COUNT(CASE WHEN b.status = 'cancelled' THEN 1 END)::numeric / COUNT(b.id) * 100, 2)
            ELSE 0 
        END AS cancellation_rate
    FROM services s
    LEFT JOIN bookings b ON b.service_id = s.id
        AND ($1::date IS NULL OR b.booking_date >= $1::date)
        AND ($2::date IS NULL OR b.booking_date <= $2::date)
    WHERE s.box_solution = TRUE
    GROUP BY s.id, s.name
    ORDER BY 
        CASE 
            WHEN $3 = 'popularity' THEN COUNT(b.id)
            WHEN $3 = 'revenue' THEN COUNT(b.id) * 1000
            WHEN $3 = 'attendance' THEN 
                CASE 
                    WHEN COUNT(DISTINCT b.service_id) > 0 
                    THEN COUNT(b.id)::numeric / COUNT(DISTINCT b.service_id)
                    ELSE 0 
                END
            ELSE COUNT(b.id)
        END DESC
`

	getFavoriteBoxesQuery = `
    SELECT 
        s.id as box_id,
        s.name as box_name,
        COUNT(uf.user_id) as favorites_count
    FROM services s
    LEFT JOIN user_favorites uf ON uf.service_id = s.id
    WHERE s.box_solution = TRUE
    GROUP BY s.id, s.name
    ORDER BY favorites_count DESC
    LIMIT 10
`

	getUsersAnalyticsExtendedQuery = `
    WITH user_visits AS (
        SELECT 
            user_id,
            COUNT(*) as visit_count
        FROM bookings 
        WHERE status = 'confirmed'
          AND ($1::date IS NULL OR booking_date >= $1::date)
          AND ($2::date IS NULL OR booking_date <= $2::date)
        GROUP BY user_id
    ),
    visit_distribution AS (
        SELECT 
            COUNT(CASE WHEN visit_count = 1 THEN 1 END) as one_visit,
            COUNT(CASE WHEN visit_count BETWEEN 2 AND 3 THEN 1 END) as two_three_visits,
            COUNT(CASE WHEN visit_count BETWEEN 4 AND 5 THEN 1 END) as four_five_visits,
            COUNT(CASE WHEN visit_count >= 6 THEN 1 END) as six_plus_visits
        FROM user_visits
    ),
    favorite_boxes AS (
        SELECT 
            service_id as box_id,
            COUNT(*) as favorites_count
        FROM user_favorites
        GROUP BY service_id
        ORDER BY favorites_count DESC
        LIMIT 10
    )
    SELECT 
        (SELECT COUNT(*) FROM users) as total_users,
        (SELECT COUNT(*) FROM users WHERE created_at >= COALESCE($1, NOW() - INTERVAL '30 days') 
         AND created_at <= COALESCE($2, NOW())) as new_users,
        (SELECT COUNT(DISTINCT user_id) FROM bookings 
         WHERE ($1::date IS NULL OR booking_date >= $1::date)
           AND ($2::date IS NULL OR booking_date <= $2::date)) as active_users,
        (SELECT one_visit FROM visit_distribution) as one_visit_dist,
        (SELECT two_three_visits FROM visit_distribution) as two_three_visits_dist,
        (SELECT four_five_visits FROM visit_distribution) as four_five_visits_dist,
        (SELECT six_plus_visits FROM visit_distribution) as six_plus_visits_dist
`

	getDashboardAnalyticsQuery = `
    WITH RECURSIVE date_series AS (
        SELECT $1::date as date
        UNION ALL
        SELECT date + INTERVAL '1 day'
        FROM date_series
        WHERE date < $2::date
    ),
    events_by_date AS (
        SELECT 
            ds.date,
            COUNT(DISTINCT b.service_id) as total_events,
            COUNT(b.id) as total_bookings,
            CASE 
                WHEN COUNT(b.id) > 0 
                THEN ROUND(COUNT(CASE WHEN b.status = 'confirmed' THEN 1 END)::numeric / COUNT(b.id) * 100, 2)
                ELSE 0 
            END as attendance_rate
        FROM date_series ds
        LEFT JOIN bookings b ON b.booking_date >= ds.date AND b.booking_date < ds.date + INTERVAL '1 day'
        GROUP BY ds.date
        ORDER BY ds.date
    ),
    popular_boxes AS (
        SELECT 
            s.id as box_id,
            s.name as box_name,
            COUNT(b.id) as bookings_count
        FROM services s
        LEFT JOIN bookings b ON b.service_id = s.id
            AND b.booking_date >= $1::date
            AND b.booking_date <= $2::date
        WHERE s.box_solution = TRUE
        GROUP BY s.id, s.name
        ORDER BY bookings_count DESC
        LIMIT 10
    ),
    user_stats AS (
        SELECT 
            (SELECT COUNT(*) FROM users) as total,
            (SELECT COUNT(*) FROM users 
             WHERE created_at >= $1::date) as new_this_period,
            (SELECT COUNT(DISTINCT user_id) FROM bookings 
             WHERE booking_date >= $1::date) as active_this_period
    )
    SELECT 
        COALESCE(json_agg(row_to_json(events_by_date)) FILTER (WHERE events_by_date.date IS NOT NULL), '[]') as events_data,
        COALESCE(json_agg(row_to_json(popular_boxes)) FILTER (WHERE popular_boxes.box_id IS NOT NULL), '[]') as boxes_data,
        (SELECT row_to_json(user_stats) FROM user_stats) as user_stats_data
`
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

func (r *AnalyticsRepo) GetOverviewAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsOverview, error) {
	const operation = "get_overview_analytics"

	var result struct {
		TotalEvents       int64   `db:"total_events"`
		TotalBookings     int64   `db:"total_bookings"`
		TotalUsers        int64   `db:"total_users"`
		AttendanceRate    float64 `db:"attendance_rate"`
		AverageAttendance float64 `db:"average_attendance"`
		Revenue 		  int64   `db:"revenue"`
	}

	err := repository.WithDBMetrics(operation, func() error {
		return r.db.GetContext(ctx, &result, getAnalyticsOverviewQuery, dateFrom, dateTo)
	})

	if err != nil {
		return dto.AnalyticsOverview{}, err
	}

	period := dto.DateRange{}
	if dateFrom != nil {
		period.From = dateFrom.Format("2006-01-02")
	}
	if dateTo != nil {
		period.To = dateTo.Format("2006-01-02")
	}

	return dto.AnalyticsOverview{
		Period:            period,
		TotalEvents:       result.TotalEvents,
		TotalBookings:     result.TotalBookings,
		TotalUsers:        result.TotalUsers,
		AttendanceRate:    result.AttendanceRate,
		AverageAttendance: result.AverageAttendance,
		Revenue:           result.Revenue,
	}, nil
}

func (r *AnalyticsRepo) GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error) {
	const operation = "get_boxes_analytics_extended"
	var rows []dto.AnalyticsBoxItem

	var fromParam, toParam *time.Time
	if dateFrom != nil {
		fromParam = dateFrom
	}
	if dateTo != nil {
		toParam = dateTo
	}

	_, err := repository.WithDBMetricsValue(operation, func() ([]dto.AnalyticsBoxItem, error) {
		err := r.db.SelectContext(ctx, &rows, getBoxesAnalyticsExtendedQuery, fromParam, toParam, sortBy)
		return rows, err
	})

	return rows, err
}

func (r *AnalyticsRepo) GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error) {
	const operation = "get_users_analytics_extended"

	var result struct {
		TotalUsers         int64 `db:"total_users"`
		NewUsers           int64 `db:"new_users"`
		ActiveUsers        int64 `db:"active_users"`
		OneVisitDist       int64 `db:"one_visit_dist"`
		TwoThreeVisitsDist int64 `db:"two_three_visits_dist"`
		FourFiveVisitsDist int64 `db:"four_five_visits_dist"`
		SixPlusVisitsDist  int64 `db:"six_plus_visits_dist"`
	}

	var fromParam, toParam *time.Time
	if dateFrom != nil {
		fromParam = dateFrom
	}
	if dateTo != nil {
		toParam = dateTo
	}

	err := repository.WithDBMetrics(operation, func() error {
		return r.db.GetContext(ctx, &result, getUsersAnalyticsExtendedQuery, fromParam, toParam)
	})

	if err != nil {
		return dto.AnalyticsUsers{}, err
	}

	var favoriteBoxes []dto.FavoriteBoxItem
	err = repository.WithDBMetrics(operation+"_favorite_boxes", func() error {
		return r.db.SelectContext(ctx, &favoriteBoxes, getFavoriteBoxesQuery)
	})

	if err != nil {
		return dto.AnalyticsUsers{}, err
	}

	return dto.AnalyticsUsers{
		TotalUsers:  result.TotalUsers,
		NewUsers:    result.NewUsers,
		ActiveUsers: result.ActiveUsers,
		AttendanceDistribution: dto.AttendanceDistribution{
			OneVisit:       result.OneVisitDist,
			TwoThreeVisits: result.TwoThreeVisitsDist,
			FourFiveVisits: result.FourFiveVisitsDist,
			SixPlusVisits:  result.SixPlusVisitsDist,
		},
		FavoriteBoxes: favoriteBoxes,
	}, nil
}

func (r *AnalyticsRepo) GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error) {
	const operation = "get_dashboard_analytics"

	var fromParam, toParam *time.Time
	if dateFrom != nil {
		fromParam = dateFrom
	}
	if dateTo != nil {
		toParam = dateTo
	}

	var result struct {
		EventsData json.RawMessage `db:"events_data"`
		BoxesData  json.RawMessage `db:"boxes_data"`
		UserStats  json.RawMessage `db:"user_stats_data"`
	}

	err := repository.WithDBMetrics(operation, func() error {
		return r.db.GetContext(ctx, &result, getDashboardAnalyticsQuery, fromParam, toParam)
	})

	if err != nil {
		return dto.AnalyticsDashboard{}, err
	}

	var dashboard dto.AnalyticsDashboard

	if len(result.EventsData) > 0 && string(result.EventsData) != "[]" {
		err = json.Unmarshal(result.EventsData, &dashboard.EventsByDate)
		if err != nil {
			return dto.AnalyticsDashboard{}, err
		}
	}

	if len(result.BoxesData) > 0 && string(result.BoxesData) != "[]" {
		err = json.Unmarshal(result.BoxesData, &dashboard.PopularBoxes)
		if err != nil {
			return dto.AnalyticsDashboard{}, err
		}
	}

	if len(result.UserStats) > 0 {
		err = json.Unmarshal(result.UserStats, &dashboard.UserStats)
		if err != nil {
			return dto.AnalyticsDashboard{}, err
		}
	}

	return dashboard, nil
}
