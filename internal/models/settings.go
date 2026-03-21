package models

import "database/sql"

// SettingRow — строка таблицы settings (key, value, category).
type SettingRow struct {
	Category string         `db:"category"`
	Key      sql.NullString `db:"key"`
	Value    sql.NullString `db:"value"`
}

// Setting — строка таблицы settings (key, value, category).
type Setting struct {
	Category string `db:"category"`
	Key      string `db:"key"`
	Value    string `db:"value"`
}

//// NotificationsSettings — настройки уведомлений (API настроек).
//type NotificationsSettings struct {
//	TelegramBotToken    *string `db:"telegram_bot_token"`
//	AutoReminders       *bool   `db:"auto_reminders"`
//	ReminderHoursBefore *int    `db:"reminder_hours_before"`
//}
//
//// BookingSettings — настройки бронирования (API настроек).
//type BookingSettings struct {
//	MaxSlotsPerEvent         *int  `db:"max_slots_per_event"`
//	AllowOverbooking         *bool `db:"allow_overbooking"`
//	CancellationAllowedHours *int  `db:"cancellation_allowed_hours"`
//}
//
//// GeneralSettings — общие настройки (API настроек).
//type GeneralSettings struct {
//	SiteName     *string `db:"site_name"`
//	ContactEmail *string `db:"contact_email"`
//	ContactPhone *string `db:"contact_phone"`
//}
//
//// Settings — агрегат настроек для API.
//type Settings struct {
//	Notifications NotificationsSettings `db:"notifications"`
//	Booking       BookingSettings       `db:"booking"`
//	General       GeneralSettings       `db:"general"`
//}
//
//type SettingsUpdateRequest struct {
//	Notifications NotificationsSettings `db:"notifications"`
//	Booking       BookingSettings       `db:"booking"`
//	General       GeneralSettings       `db:"general"`
//}
