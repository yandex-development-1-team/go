package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

type SettingsRep struct {
	client *sqlx.DB
}

func NewSettingsRep(client *sqlx.DB) *SettingsRep {
	return &SettingsRep{client: client}
}

func (r *SettingsRep) GetSettings(ctx context.Context) (models.SettingsFormMessages, error) {
	var settings models.SettingsFormMessages

	query := `
		SELECT 
			MAX(value) FILTER (WHERE key = 'welcome_message') AS welcome_message,
			MAX(value) FILTER (WHERE key = 'record_confirmation') AS record_confirmation,
			MAX(value) FILTER (WHERE key = 'event_reminder_for_week') AS event_reminder_for_week,
			MAX(value) FILTER (WHERE key = 'event_reminder_for_24_hours') AS event_reminder_for_24_hours,
			MAX(value) FILTER (WHERE key = 'cancellation_message') AS cancellation_message,
			MAX(value) FILTER (WHERE key = 'thanks_message') AS thanks_message,
			MAX(value) FILTER (WHERE key = 'system_err_message') AS system_err_message
		FROM settings_messages;
	`

	err := r.client.GetContext(ctx, &settings, query)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("failed to get settings from db", zap.Error(err))
			return models.SettingsFormMessages{}, fmt.Errorf("settings messages not found in database")
		}

		logger.Error("failed to get settings from db", zap.Error(err))
		return models.SettingsFormMessages{}, fmt.Errorf("failed to get settings messages: %w", err)
	}

	return settings, nil
}

func (r *SettingsRep) PutSettings(ctx context.Context, newSettings models.SettingsFormMessages) error {
	query := `
        INSERT INTO settings_messages (key, value)
        VALUES 
            ('welcome_message', :welcome_message),
            ('record_confirmation', :record_confirmation),
            ('event_reminder_for_week', :event_reminder_for_week),
            ('event_reminder_for_24_hours', :event_reminder_for_24_hours),
            ('cancellation_message', :cancellation_message),
            ('thanks_message', :thanks_message),
            ('system_err_message', :system_err_message)
        ON CONFLICT (key) 
        DO UPDATE SET 
            value = EXCLUDED.value,
            updated_at = NOW();`

	_, err := r.client.NamedExecContext(ctx, query, newSettings)

	if err != nil {
		logger.Error("failed to update settings messages from DB", zap.Error(err))
		return err
	}

	return nil
}

func (r *SettingsRep) PostSettings(ctx context.Context, newSettings models.SettingsPermissions) error {
	query := `
        INSERT INTO role_permissions (role, permissions) 
        VALUES ($1, $2)
        ON CONFLICT (role) 
        DO UPDATE SET permissions = EXCLUDED.permissions
    `

	_, err := r.client.ExecContext(ctx, query, newSettings.Role, pq.Array(newSettings.Permissions))

	return err
}

//const getSettingsQuery = `SELECT key, value, category FROM settings`
//
//func (r *SettingsRep) GetSettings(ctx context.Context) ([]models.SettingRow, error) {
//	var settings []models.SettingRow
//	err := r.client.SelectContext(ctx, &settings, getSettingsQuery)
//	if err != nil {
//		logger.Error("failed to get settings from db", zap.Error(err))
//		return nil, err
//	}
//
//	if len(settings) == 0 {
//		logger.Error("response settings is empty from repository")
//		return settings, fmt.Errorf("response settings is empty from repository")
//	}
//
//	return settings, nil
//}
//
//func (r *SettingsRep) PutSettings(ctx context.Context, newSettings []models.Setting) (time.Time, error) {
//	if len(newSettings) == 0 {
//		logger.Error("request settings is empty from repository")
//		return time.Now(), fmt.Errorf("request settings is empty from repository")
//	}
//
//	// Подготовка массивов для bulk insert
//	categories := make([]string, len(newSettings))
//	keys := make([]string, len(newSettings))
//	values := make([]string, len(newSettings))
//
//	for i, s := range newSettings {
//		categories[i] = s.Category
//		keys[i] = s.Key
//		values[i] = s.Value
//	}
//
//	query := `
//        INSERT INTO settings (category, key, value, updated_at)
//        SELECT
//            unnest($1::varchar[])::setting_category_type AS category,
//            unnest($2::varchar[]) AS key,
//            unnest($3::text[]) AS value,
//            $4::timestamptz AS updated_at
//        ON CONFLICT (category, key)
//        DO UPDATE SET
//            value = EXCLUDED.value,
//            updated_at = EXCLUDED.updated_at
//    `
//
//	now := time.Now()
//	_, err := r.client.ExecContext(ctx, query,
//		pq.Array(categories),
//		pq.Array(keys),
//		pq.Array(values),
//		now,
//	)
//	if err != nil {
//		logger.Error("failed to update settings from DB", zap.Error(err))
//		return time.Time{}, err
//	}
//
//	return now, nil
//}
