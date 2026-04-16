package dto

import "time"

type ExportType string

const (
	ExportTypeBoxes ExportType = "boxes"
	ExportTypeUsers ExportType = "users"
)

type ExportFormat string

const (
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatCSV  ExportFormat = "csv"
)

type AnalyticsExportRequest struct {
	Type     ExportType
	DateFrom *time.Time
	DateTo   *time.Time
	Format   ExportFormat
}

type AnalyticsBoxRow struct {
	ServiceID         int64   `db:"service_id"`
	ServiceName       string  `db:"service_name"`
	TotalBookings     int64   `db:"total_bookings"`
	ConfirmedBookings int64   `db:"confirmed_bookings"`
	CancelledBookings int64   `db:"cancelled_bookings"`
	CancellationRate  float64 `db:"cancellation_rate"`
}

type AnalyticsUserRow struct {
	UserID        int64     `db:"user_id"`
	FirstName     string    `db:"first_name"`
	LastName      string    `db:"last_name"`
	Email         string    `db:"email"`
	TotalBookings int64     `db:"total_bookings"`
	RegisteredAt  time.Time `db:"registered_at"`
}

type DateRange struct {
	From string `json:"from"` // формат: "YYYY-MM-DD"
	To   string `json:"to"`   // формат: "YYYY-MM-DD"
}

type AnalyticsOverview struct {
	Period            DateRange `json:"period"`
	TotalEvents       int64     `json:"total_events"`
	TotalBookings     int64     `json:"total_bookings"`
	TotalUsers        int64     `json:"total_users"`
	AttendanceRate    float64   `json:"attendance_rate"`
	AverageAttendance float64   `json:"average_attendance"`
	Revenue           int64     `json:"revenue"`
}

type AnalyticsBoxItem struct {
	BoxID             int64   `db:"box_id" json:"box_id"`
	BoxName           string  `db:"box_name" json:"box_name"`
	TotalEvents       int64   `db:"total_events" json:"total_events"`
	TotalBookings     int64   `db:"total_bookings" json:"total_bookings"`
	AttendanceRate    float64 `db:"attendance_rate" json:"attendance_rate"`
	AverageAttendance float64 `db:"average_attendance" json:"average_attendance"`
	ConfirmedBookings int64   `db:"confirmed_bookings" json:"confirmed_bookings"`
	CancelledBookings int64   `db:"cancelled_bookings" json:"cancelled_bookings"`
	Revenue           int64   `db:"revenue" json:"revenue"`
	CancellationRate  float64 `db:"cancellation_rate" json:"cancellation_rate"`
}

type AnalyticsBoxesResponse struct {
	Boxes []AnalyticsBoxItem `json:"boxes"`
}

type AttendanceDistribution struct {
	OneVisit       int64 `json:"1_visit"`
	TwoThreeVisits int64 `json:"2_3_visits"`
	FourFiveVisits int64 `json:"4_5_visits"`
	SixPlusVisits  int64 `json:"6_plus_visits"`
}

type FavoriteBoxItem struct {
	BoxID          int64  `json:"box_id"`
	BoxName        string `json:"box_name"`
	FavoritesCount int64  `json:"favorites_count"`
}

type AnalyticsUsers struct {
	TotalUsers             int64                  `json:"total_users"`
	NewUsers               int64                  `json:"new_users"`
	ActiveUsers            int64                  `json:"active_users"`
	AttendanceDistribution AttendanceDistribution `json:"attendance_distribution"`
	FavoriteBoxes          []FavoriteBoxItem      `json:"favorite_boxes"`
}

type EventDateStat struct {
	Date           string  `json:"date"` // формат: "YYYY-MM-DD"
	TotalEvents    int64   `json:"total_events"`
	TotalBookings  int64   `json:"total_bookings"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type PopularBox struct {
	BoxID         int64  `json:"box_id"`
	BoxName       string `json:"box_name"`
	BookingsCount int64  `json:"bookings_count"`
}

type UserStatistics struct {
	Total            int64 `json:"total"`
	NewThisPeriod    int64 `json:"new_this_period"`
	ActiveThisPeriod int64 `json:"active_this_period"`
}

type AnalyticsDashboard struct {
	EventsByDate []EventDateStat `json:"events_by_date"`
	PopularBoxes []PopularBox    `json:"popular_boxes"`
	UserStats    UserStatistics  `json:"user_stats"`
}
