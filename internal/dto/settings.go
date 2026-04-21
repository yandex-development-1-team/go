package dto

type SettingsFormMessages struct {
	WelcomeMessage          string `json:"welcome_message"`
	RecordConfirmation      string `json:"record_confirmation"`
	EventReminderForWeek    string `json:"event_reminder_for_week"`
	EventReminderFor24Hours string `json:"event_reminder_for_24_hours"`
	CancellationMessage     string `json:"cancellation_message"`
	ThanksMessage           string `json:"thanks_message"`
	SystemErrMessage        string `json:"system_err_message"`
}

type SettingsResponse struct {
	Notifications Notifications `json:"notifications,omitempty"`
	Booking       Booking       `json:"booking"`
	General       General       `json:"general"`
}

type SettingsRequest struct {
	Notifications Notifications `json:"notifications"`
	Booking       Booking       `json:"booking"`
	General       General       `json:"general"`
}

type Notifications struct {
	TelegramBotToken    *string `json:"telegram_bot_token,omitempty"`
	AutoReminders       *bool   `json:"auto_reminders,omitempty"`
	ReminderHoursBefore *int    `json:"reminder_hours_before,omitempty"`
}

type Booking struct {
	MaxSlotsPerEvent         *int  `json:"max_slots_per_event,omitempty"`
	AllowOverbooking         *bool `json:"allow_overbooking,omitempty"`
	CancellationAllowedHours *int  `json:"cancellation_allowed_hours,omitempty"`
}

type General struct {
	SiteName     *string `json:"site_name,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty"`
	ContactPhone *string `json:"contact_phone,omitempty"`
}

type SettingsPermissionsRequest struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}
