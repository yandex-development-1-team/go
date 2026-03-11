package models

import "database/sql"

// Setting — строка таблицы settings (key, value, category).
type Setting struct {
	Category string         `db:"category"`
	Key      sql.NullString `db:"key"`
	Value    sql.NullString `db:"value"`
}

// SettingsNotifications — настройки уведомлений (API настроек).
type SettingsNotifications struct {
	TelegramBotToken    string
	AutoReminders       bool
	ReminderHoursBefore int
}

// SettingsBooking — настройки бронирования (API настроек).
type SettingsBooking struct {
	MaxSlotsPerEvent         int
	AllowOverbooking         bool
	CancellationAllowedHours int
}

// SettingsGeneral — общие настройки (API настроек).
type SettingsGeneral struct {
	SiteName     string
	ContactEmail string
	ContactPhone string
}

// Settings — агрегат настроек для API.
type Settings struct {
	Notifications SettingsNotifications
	Booking       SettingsBooking
	General       SettingsGeneral
}
