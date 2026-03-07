package postgres

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/repository/postgres/models"
	"go.uber.org/zap"
)

type SettingsRep struct {
	client *sqlx.DB
}

func NewSettingsRep(client *sqlx.DB) *SettingsRep {
	return &SettingsRep{client: client}
}

func (r *SettingsRep) GetSettings(ctx context.Context) ([]models.Setting, error) {
	query := `SELECT key, value, category FROM settings`

	var settings []models.Setting

	err := r.client.SelectContext(ctx, &settings, query)
	if err != nil {
		logger.Error("failed to get settings from db", zap.Error(err))
		return settings, err
	}

	return settings, nil
}
