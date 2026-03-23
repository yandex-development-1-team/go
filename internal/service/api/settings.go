package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
	serviceModels "github.com/yandex-development-1-team/go/internal/service/api/models"
)

type SettingsService struct {
	settingsRepo repository.SettingsRepository
}

func NewSettingsService(settingsRepo repository.SettingsRepository) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo}
}

func (a SettingsService) GetSettings(ctx context.Context) ([]serviceModels.Setting, error) {
	settingsDB, err := a.settingsRepo.GetSettings(ctx)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return []serviceModels.Setting{}, err
	}

	settings := convertRespDBToRespService(settingsDB)

	return settings, nil
}

func (a SettingsService) PutSettings(ctx context.Context, reqService []serviceModels.Setting) (time.Time, error) {
	if len(reqService) == 0 {
		logger.Error("request settings is empty from repository")
		return time.Now(), fmt.Errorf("request settings is empty from repository")
	}

	reqBD := convertReqServiceToReqBD(reqService)

	updatedAt, err := a.settingsRepo.PutSettings(ctx, reqBD)
	if err != nil {
		logger.Error("failed to get settings from service", zap.Error(err))
		return updatedAt, fmt.Errorf("failed to get settings from service: %w", err)
	}

	return updatedAt, nil
}

func convertReqServiceToReqBD(reqService []serviceModels.Setting) []models.Setting {
	var reqBD []models.Setting

	for _, setting := range reqService {
		reqBD = append(reqBD, models.Setting{
			Category: setting.Category,
			Key:      setting.Key,
			Value:    setting.Value,
		})
	}

	return reqBD
}

func convertRespDBToRespService(reqService []models.SettingRow) []serviceModels.Setting {
	var respService []serviceModels.Setting

	for _, setting := range reqService {
		respService = append(respService, serviceModels.Setting{
			Category: setting.Category,
			Key:      setting.Key.String,
			Value:    setting.Value.String,
		})
	}

	return respService
}
