package models

// Setting — строка таблицы settings (key, value, category).
type Setting struct {
	Category string
	Key      string
	Value    string
}

//type Notifications struct {
//	TelegramBotToken    *string
//	AutoReminders       *bool
//	ReminderHoursBefore *int
//}
//
//type Booking struct {
//	MaxSlotsPerEvent         *int
//	AllowOverbooking         *bool
//	CancellationAllowedHours *int
//}
//
//type General struct {
//	SiteName     *string
//	ContactEmail *string
//	ContactPhone *string
//}
//
//type Settings struct {
//	Notifications Notifications
//	Booking       Booking
//	General       General
//}
//
//type SettingsUpdateRequest struct {
//	Notifications Notifications
//	Booking       Booking
//	General       General
//}
