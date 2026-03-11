package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
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
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "failed to get settings",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}
