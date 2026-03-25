package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

const getSettingsQuery = `
	SELECT key, value, category FROM settings`

type SettingsRep struct {
	client *sqlx.DB
}

func NewSettingsRep(client *sqlx.DB) *SettingsRep {
	return &SettingsRep{client: client}
}

func (r *SettingsRep) GetSettings(ctx context.Context) ([]models.SettingRow, error) {
	var settings []models.SettingRow
	err := r.client.SelectContext(ctx, &settings, getSettingsQuery)
	if err != nil {
		logger.Error("failed to get settings from db", zap.Error(err))
		return nil, err
	}

	if len(settings) == 0 {
		logger.Error("response settings is empty from repository")
		return settings, fmt.Errorf("response settings is empty from repository")
	}

	return settings, nil
}

func (r *SettingsRep) PutSettings(ctx context.Context, newSettings []models.Setting) (time.Time, error) {
	if len(newSettings) == 0 {
		logger.Error("request settings is empty from repository")
		return time.Now(), fmt.Errorf("request settings is empty from repository")
	}

	// Подготовка массивов для bulk insert
	categories := make([]string, len(newSettings))
	keys := make([]string, len(newSettings))
	values := make([]string, len(newSettings))

	for i, s := range newSettings {
		categories[i] = s.Category
		keys[i] = s.Key
		values[i] = s.Value
	}

	query := `
        INSERT INTO settings (category, key, value, updated_at)
        SELECT 
            unnest($1::varchar[])::setting_category_type AS category,
            unnest($2::varchar[]) AS key,
            unnest($3::text[]) AS value,
            $4::timestamptz AS updated_at
        ON CONFLICT (category, key)
        DO UPDATE SET
            value = EXCLUDED.value,
            updated_at = EXCLUDED.updated_at
    `

	now := time.Now()
	_, err := r.client.ExecContext(ctx, query,
		pq.Array(categories),
		pq.Array(keys),
		pq.Array(values),
		now,
	)
	if err != nil {
		logger.Error("failed to update settings from DB", zap.Error(err))
		return time.Time{}, err
	}

	return now, nil
}
