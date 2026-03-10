package service

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/service/api/models"
)

// Категории настроек (хранятся в БД)
const (
	CategoryNotifications = "notifications"
	CategoryBooking       = "booking"
	CategoryGeneral       = "general"
)

// Ключи раздела notifications
const (
	KeyTelegramBotToken    = "telegram_bot_token"
	KeyAutoReminders       = "auto_reminders"
	KeyReminderHoursBefore = "reminder_hours_before"
)

// Ключи раздела booking
const (
	KeyMaxSlotsPerEvent         = "max_slots_per_event"
	KeyAllowOverbooking         = "allow_overbooking"
	KeyCancellationAllowedHours = "cancellation_allowed_hours"
)

// Ключи раздела general
const (
	KeySiteName     = "site_name"
	KeyContactEmail = "contact_email"
	KeyContactPhone = "contact_phone"
)

type SettingsService struct {
	settingsRepo repository.SettingsRepository
}

func NewSettingsService(settingsRepo repository.SettingsRepository) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo}
}

func (a SettingsService) GetSettings(ctx context.Context) (models.Settings, error) {
	var settings models.Settings

	settingsDB, err := a.settingsRepo.GetSettings(ctx)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return settings, err
	}

	for _, row := range settingsDB {
		switch row.Category {
		case CategoryNotifications:
			if err := mapNotification(&settings.Notifications, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		case CategoryBooking:
			if err := mapBooking(&settings.Booking, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		case CategoryGeneral:
			if err := mapGeneral(&settings.General, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		}
	}

	return settings, nil
}

func mapNotification(n *models.Notifications, key, value string) error {
	switch key {
	case KeyTelegramBotToken:
		n.TelegramBotToken = value
	case KeyAutoReminders:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %w", KeyAutoReminders, err)
		}
		n.AutoReminders = val
	case KeyReminderHoursBefore:
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %w", KeyReminderHoursBefore, err)
		}
		n.ReminderHoursBefore = val
	}
	return nil
}

func mapBooking(b *models.Booking, key, value string) error {
	switch key {
	case KeyMaxSlotsPerEvent:
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %w", KeyMaxSlotsPerEvent, err)
		}
		b.MaxSlotsPerEvent = val
	case KeyAllowOverbooking:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %w", KeyAllowOverbooking, err)
		}
		b.AllowOverbooking = val
	case KeyCancellationAllowedHours:
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %w", KeyCancellationAllowedHours, err)
		}
		b.CancellationAllowedHours = val
	}
	return nil
}

func mapGeneral(g *models.General, key, value string) error {
	switch key {
	case KeySiteName:
		g.SiteName = value
	case KeyContactEmail:
		g.ContactEmail = value
	case KeyContactPhone:
		g.ContactPhone = value
	}
	return nil
}
