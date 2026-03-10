package dto

// SettingsResponse — ответ GET /api/v1/settings (docs/openapi.json Settings).
type SettingsResponse struct {
	Notifications SettingsNotifications `json:"notifications"`
	Booking       SettingsBooking       `json:"booking"`
	General       SettingsGeneral       `json:"general"`
}

// SettingsNotifications — настройки уведомлений.
type SettingsNotifications struct {
	TelegramBotToken    string `json:"telegram_bot_token"`
	AutoReminders       bool   `json:"auto_reminders"`
	ReminderHoursBefore int    `json:"reminder_hours_before"`
}

// SettingsBooking — настройки бронирования.
type SettingsBooking struct {
	MaxSlotsPerEvent         int  `json:"max_slots_per_event"`
	AllowOverbooking         bool `json:"allow_overbooking"`
	CancellationAllowedHours int  `json:"cancellation_allowed_hours"`
}

// SettingsGeneral — общие настройки.
type SettingsGeneral struct {
	SiteName     string `json:"site_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}
