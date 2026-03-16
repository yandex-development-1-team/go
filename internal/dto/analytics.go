package dto

import "time"

// ExportType defines the analytics entity to export.
type ExportType string

const (
	ExportTypeBoxes ExportType = "boxes"
	ExportTypeUsers ExportType = "users"
)

// ExportFormat defines the output file format.
type ExportFormat string

const (
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatCSV  ExportFormat = "csv"
)

// AnalyticsExportRequest contains parsed and validated export query parameters.
type AnalyticsExportRequest struct {
	Type     ExportType
	DateFrom *time.Time
	DateTo   *time.Time
	Format   ExportFormat
}

// AnalyticsBoxRow represents one box (service) with its booking aggregates.
type AnalyticsBoxRow struct {
	ServiceID         int64   `db:"service_id"`
	ServiceName       string  `db:"service_name"`
	TotalBookings     int64   `db:"total_bookings"`
	ConfirmedBookings int64   `db:"confirmed_bookings"`
	CancelledBookings int64   `db:"cancelled_bookings"`
	CancellationRate  float64 `db:"cancellation_rate"`
}

// AnalyticsUserRow represents one user with their booking count.
type AnalyticsUserRow struct {
	UserID        int64     `db:"user_id"`
	FirstName     string    `db:"first_name"`
	LastName      string    `db:"last_name"`
	Email         string    `db:"email"`
	TotalBookings int64     `db:"total_bookings"`
	RegisteredAt  time.Time `db:"registered_at"`
}
