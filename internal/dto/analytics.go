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
