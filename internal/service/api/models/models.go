package models

type Notifications struct {
	TelegramBotToken    string `json:"telegram_bot_token"`
	AutoReminders       bool   `json:"auto_reminders"`
	ReminderHoursBefore int    `json:"reminder_hours_before"`
}

type Booking struct {
	MaxSlotsPerEvent         int  `json:"max_slots_per_event"`
	AllowOverbooking         bool `json:"allow_overbooking"`
	CancellationAllowedHours int  `json:"cancellation_allowed_hours"`
}

type General struct {
	SiteName     string `json:"site_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}

type Settings struct {
	Notifications Notifications
	Booking       Booking
	General       General
}

type SettingsUpdateRequest struct {
	Notifications Notifications `json:"notifications"`
	Booking       Booking       `json:"booking"`
	General       General       `json:"general"`
}
