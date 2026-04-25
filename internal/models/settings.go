package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type SettingRow struct {
	Category string         `db:"category"`
	Key      sql.NullString `db:"key"`
	Value    sql.NullString `db:"value"`
}

type Setting struct {
	Category string `db:"category"`
	Key      string `db:"key"`
	Value    string `db:"value"`
}

type SettingsPermissions struct {
	Role        string         `db:"role"`
	Permissions pq.StringArray `db:"permissions"`
}

type SettingsFormMessages struct {
	WelcomeMessage          string `db:"welcome_message"`
	RecordConfirmation      string `db:"record_confirmation"`
	EventReminderForWeek    string `db:"event_reminder_for_week"`
	EventReminderFor24Hours string `db:"event_reminder_for_24_hours"`
	CancellationMessage     string `db:"cancellation_message"`
	ThanksMessage           string `db:"thanks_message"`
	SystemErrMessage        string `db:"system_err_message"`
}
