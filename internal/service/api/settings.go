package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type SettingsService struct {
	settingsRepo repository.SettingsRepository
}

func NewSettingsService(settingsRepo repository.SettingsRepository) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo}
}

func (a SettingsService) GetSettings(ctx context.Context) ([]models.Setting, error) {
	settingsDB, err := a.settingsRepo.GetSettings(ctx)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return []models.Setting{}, err
	}

	settings := convertRespDBToRespService(settingsDB)

	return settings, nil
}

func (a SettingsService) PutSettings(ctx context.Context, reqService []models.Setting) (time.Time, error) {
	if len(reqService) == 0 {
		logger.Error("request settings is empty from repository")
		return time.Now(), fmt.Errorf("request settings is empty from repository")
	}

	updatedAt, err := a.settingsRepo.PutSettings(ctx, reqService)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return updatedAt, fmt.Errorf("failed to get settings from service: %w", err)
	}

	return updatedAt, nil
}

func (a SettingsService) PostSettings(ctx context.Context, reqService models.SettingsPermissions) error {
	err := a.settingsRepo.PostSettings(ctx, reqService)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return fmt.Errorf("failed to get settings from service: %w", err)
	}

	return nil
}

func convertRespDBToRespService(reqService []models.SettingRow) []models.Setting {
	var respService []models.Setting

	for _, setting := range reqService {
		respService = append(respService, models.Setting{
			Category: setting.Category,
			Key:      setting.Key.String,
			Value:    setting.Value.String,
		})
	}

	return respService
}
