package handlers

import (
	serviceModels "github.com/yandex-development-1-team/go/internal/service/api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	api "github.com/yandex-development-1-team/go/internal/service/api"
)

type SettingsHandler struct {
	service *api.SettingsService
}

func NewSettingsHandler(service *api.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (a SettingsHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	settings, err := a.service.GetSettings(ctx)
	if err != nil {
		logger.Error("failed to get settings from handler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get settings",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (a SettingsHandler) Put(c *gin.Context) {
	ctx := c.Request.Context()
	var req serviceModels.SettingsUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to get settings from put request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": models.ErrValidation,
		})
	}

	updatedAt, err := a.service.PutSettings(ctx, req)
	if err != nil {
		logger.Error("failed to get settings from handler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "successful",
		"updated_at": updatedAt,
	})
}
