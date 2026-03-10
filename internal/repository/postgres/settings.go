package repository

import (
	"context"

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

func (r *SettingsRep) GetSettings(ctx context.Context) ([]models.Setting, error) {
	var settings []models.Setting
	err := r.client.SelectContext(ctx, &settings, getSettingsQuery)
	if err != nil {
		logger.Error("failed to get settings from db", zap.Error(err))
		return nil, err
	}
	return settings, nil
}
