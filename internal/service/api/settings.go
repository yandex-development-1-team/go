package api

import (
	"context"
	"fmt"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/repository/postgres"
	"github.com/yandex-development-1-team/go/internal/service/api/models"
	"go.uber.org/zap"
	"strconv"
)

type SettingsService struct {
	dbClient *postgres.SettingsRep
}

func NewSettingsService(dbClient *postgres.SettingsRep) *SettingsService {
	return &SettingsService{dbClient: dbClient}
}

func (a SettingsService) GetSettings(ctx context.Context) (models.Settings, error) {
	var settings models.Settings

	settingsDB, err := a.dbClient.GetSettings(ctx)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return settings, err
	}

	for _, row := range settingsDB {
		switch row.Category {
		case "notifications":
			if err := mapNotification(&settings.Notifications, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		case "booking":
			if err := mapBooking(&settings.Booking, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		case "general":
			if err := mapGeneral(&settings.General, row.Key.String, row.Value.String); err != nil {
				return settings, err
			}
		}
	}

	return settings, nil
}

func mapNotification(n *models.Notifications, key, value string) error {
	switch key {
	case "telegram_bot_token":
		n.TelegramBotToken = value
	case "auto_reminders":
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for auto_reminders: %w", err)
		}
		n.AutoReminders = val
	case "reminder_hours_before":
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for reminder_hours_before: %w", err)
		}
		n.ReminderHoursBefore = val
	}
	return nil
}

func mapBooking(b *models.Booking, key, value string) error {
	switch key {
	case "max_slots_per_event":
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for max_slots_per_event: %w", err)
		}
		b.MaxSlotsPerEvent = val
	case "allow_overbooking":
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for allow_overbooking: %w", err)
		}
		b.AllowOverbooking = val
	case "cancellation_allowed_hours":
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for cancellation_allowed_hours: %w", err)
		}
		b.CancellationAllowedHours = val
	}
	return nil
}

func mapGeneral(g *models.General, key, value string) error {
	switch key {
	case "site_name":
		g.SiteName = value
	case "contact_email":
		g.ContactEmail = value
	case "contact_phone":
		g.ContactPhone = value
	}
	return nil
}
