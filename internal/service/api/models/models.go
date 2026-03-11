package models

type Notifications struct {
	TelegramBotToken    string
	AutoReminders       bool
	ReminderHoursBefore int
}

type Booking struct {
	MaxSlotsPerEvent         int
	AllowOverbooking         bool
	CancellationAllowedHours int
}

type General struct {
	SiteName     string
	ContactEmail string
	ContactPhone string
}

type Settings struct {
	Notifications Notifications
	Booking       Booking
	General       General
}
