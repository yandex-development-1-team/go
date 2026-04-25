package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	service "github.com/yandex-development-1-team/go/internal/service/api"
)

type SettingsHandler struct {
	service *service.SettingsService
}

func NewSettingsHandler(service *service.SettingsService) *SettingsHandler {
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

	settingsDTO := convertModelsToDTOFromSettingsMessages(settings)

	c.JSON(http.StatusOK, settingsDTO)
}

func (a SettingsHandler) GetPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	role := c.Param("role")

	if role == "" {
		logger.Error("failed to get role from URL")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get role from URL",
		})
		return
	}

	permissions, err := a.service.GetSettingsPermissions(ctx, role)
	if err != nil {
		logger.Error("failed to get settings permissions from handler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get settings permissions",
		})
		return
	}

	permissionsDTO := convertServiceToDTOFromSettingsPermissions(permissions)

	c.JSON(http.StatusOK, permissionsDTO)
}

func (a SettingsHandler) Put(c *gin.Context) {
	ctx := c.Request.Context()

	var reqDTO dto.SettingsFormMessages

	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		logger.Error("failed to get settings messages from put request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": models.ErrValidation,
		})
		return
	}

	reqService := convertDTOToModelsFromSettingsMessages(reqDTO)

	err := a.service.PutSettings(ctx, reqService)
	if err != nil {
		logger.Error("failed to get settings messages from handler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get settings messages from handler: %s", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
	})
}

func (a SettingsHandler) Post(c *gin.Context) {
	ctx := c.Request.Context()

	var reqDTO dto.SettingsPermissions

	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		logger.Error("failed to get settings permissions from post request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": models.ErrValidation,
		})
		return
	}

	err := validateSettingsPermissionsFromRequest(reqDTO)
	if err != nil {
		logger.Error("request settings permissions is not validate")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": models.ErrValidation,
		})
		return
	}

	reqService := convertDTOToServiceFromSettingsPermissions(reqDTO)

	err = a.service.PostSettings(ctx, reqService)
	if err != nil {
		logger.Error("failed to update settings permissions from handler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
	})
}

func convertDTOToServiceFromSettingsPermissions(reqDTO dto.SettingsPermissions) models.SettingsPermissions {
	return models.SettingsPermissions{
		Role:        reqDTO.Role,
		Permissions: reqDTO.Permissions,
	}
}

func convertServiceToDTOFromSettingsPermissions(permissions models.SettingsPermissions) dto.SettingsPermissions {
	return dto.SettingsPermissions{
		Role:        permissions.Role,
		Permissions: permissions.Permissions,
	}
}

func convertModelsToDTOFromSettingsMessages(settingsMes models.SettingsFormMessages) dto.SettingsFormMessages {
	return dto.SettingsFormMessages{
		WelcomeMessage:          settingsMes.WelcomeMessage,
		RecordConfirmation:      settingsMes.RecordConfirmation,
		EventReminderForWeek:    settingsMes.EventReminderForWeek,
		EventReminderFor24Hours: settingsMes.EventReminderFor24Hours,
		CancellationMessage:     settingsMes.CancellationMessage,
		ThanksMessage:           settingsMes.ThanksMessage,
		SystemErrMessage:        settingsMes.SystemErrMessage,
	}
}

func convertDTOToModelsFromSettingsMessages(settingsMes dto.SettingsFormMessages) models.SettingsFormMessages {
	return models.SettingsFormMessages{
		WelcomeMessage:          settingsMes.WelcomeMessage,
		RecordConfirmation:      settingsMes.RecordConfirmation,
		EventReminderForWeek:    settingsMes.EventReminderForWeek,
		EventReminderFor24Hours: settingsMes.EventReminderFor24Hours,
		CancellationMessage:     settingsMes.CancellationMessage,
		ThanksMessage:           settingsMes.ThanksMessage,
		SystemErrMessage:        settingsMes.SystemErrMessage,
	}
}

func validateSettingsPermissionsFromRequest(req dto.SettingsPermissions) error {
	exist := false

	for _, role := range service.Roles {
		if role == req.Role {
			exist = true
		}
	}

	if !exist {
		return fmt.Errorf("wrong role")
	}

	if len(req.Permissions) == 0 {
		return fmt.Errorf("permissions is empty")
	}

	for _, permission := range req.Permissions {
		_, ok := models.MapPermissions[permission]
		if !ok {
			return fmt.Errorf("wrong permission: %s", permission)
		}
	}

	return nil
}

//Код ниже написан по документации api, но не сходится с макетами фронта
//
//func (a SettingsHandler) Get(c *gin.Context) {
//	ctx := c.Request.Context()
//
//	settings, err := a.service.GetSettings(ctx)
//	if err != nil {
//		logger.Error("failed to get settings from handler", zap.Error(err))
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"error": "failed to get settings",
//		})
//		return
//	}
//
//	settingsDTO, err := parseSettings(settings)
//	if err != nil {
//		logger.Error("parse settings", zap.Error(err))
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"error": "failed to build settings response",
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, settingsDTO)
//}
//
//func (a SettingsHandler) Put(c *gin.Context) {
//	ctx := c.Request.Context()
//
//	var reqDTO dto.SettingsRequest
//
//	if err := c.ShouldBindJSON(&reqDTO); err != nil {
//		logger.Error("failed to get settings from put request", zap.Error(err))
//		c.JSON(http.StatusBadRequest, gin.H{
//			"error": models.ErrValidation,
//		})
//		return
//	}
//
//	reqService := prepareSettingsFromRequest(reqDTO)
//	if len(reqService) == 0 {
//		logger.Error("request settings is empty")
//		c.JSON(http.StatusBadRequest, gin.H{
//			"error": models.ErrValidation,
//		})
//		return
//	}
//
//	updatedAt, err := a.service.PutSettings(ctx, reqService)
//	if err != nil {
//		logger.Error("failed to get settings from handler", zap.Error(err))
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"error": fmt.Sprintf("failed to get settings from handler: %s", err),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message":    "successful",
//		"updated_at": updatedAt,
//	})
//}

//func parseSettings(settings []models.Setting) (*dto.SettingsResponse, error) {
//	resp := &dto.SettingsResponse{}
//
//	for _, s := range settings {
//		switch s.Category {
//		case "notifications":
//			if err := setNotificationField(&resp.Notifications, s.Key, s.Value); err != nil {
//				return nil, fmt.Errorf("notifications.%s: %w", s.Key, err)
//			}
//		case "booking":
//			if err := setBookingField(&resp.Booking, s.Key, s.Value); err != nil {
//				return nil, fmt.Errorf("booking.%s: %w", s.Key, err)
//			}
//		case "general":
//			if err := setGeneralField(&resp.General, s.Key, s.Value); err != nil {
//				return nil, fmt.Errorf("general.%s: %w", s.Key, err)
//			}
//		}
//	}
//
//	return resp, nil
//}
//
//func setNotificationField(n *dto.Notifications, key, value string) error {
//	switch key {
//	case "telegram_bot_token":
//		n.TelegramBotToken = &value
//	case "auto_reminders":
//		b, err := strconv.ParseBool(value)
//		if err != nil {
//			return err
//		}
//		n.AutoReminders = &b
//	case "reminder_hours_before":
//		i, err := strconv.Atoi(value)
//		if err != nil {
//			return err
//		}
//		n.ReminderHoursBefore = &i
//	}
//	return nil
//}
//
//func setBookingField(b *dto.Booking, key, value string) error {
//	switch key {
//	case "max_slots_per_event":
//		i, err := strconv.Atoi(value)
//		if err != nil {
//			return err
//		}
//		b.MaxSlotsPerEvent = &i
//	case "allow_overbooking":
//		bl, err := strconv.ParseBool(value)
//		if err != nil {
//			return err
//		}
//		b.AllowOverbooking = &bl
//	case "cancellation_allowed_hours":
//		i, err := strconv.Atoi(value)
//		if err != nil {
//			return err
//		}
//		b.CancellationAllowedHours = &i
//	}
//	return nil
//}
//
//func setGeneralField(g *dto.General, key, value string) error {
//	switch key {
//	case "site_name":
//		g.SiteName = &value
//	case "contact_email":
//		g.ContactEmail = &value
//	case "contact_phone":
//		g.ContactPhone = &value
//	}
//	return nil
//}
//
//func prepareSettingsFromRequest(reqDTO dto.SettingsRequest) []models.Setting {
//	var updates []models.Setting
//	// Проверяем notifications
//	if reqDTO.Notifications.TelegramBotToken != nil {
//		updates = append(updates, models.Setting{
//			Category: "notifications", Key: "telegram_bot_token", Value: *reqDTO.Notifications.TelegramBotToken})
//	}
//	if reqDTO.Notifications.AutoReminders != nil {
//		updates = append(updates, models.Setting{
//			Category: "notifications", Key: "auto_reminders", Value: fmt.Sprintf("%t", *reqDTO.Notifications.AutoReminders)})
//	}
//	if reqDTO.Notifications.ReminderHoursBefore != nil {
//		updates = append(updates, models.Setting{
//			Category: "notifications", Key: "reminder_hours_before", Value: fmt.Sprintf("%d", *reqDTO.Notifications.ReminderHoursBefore)})
//	}
//
//	// Проверяем booking
//	if reqDTO.Booking.MaxSlotsPerEvent != nil {
//		updates = append(updates, models.Setting{
//			Category: "booking", Key: "max_slots_per_event", Value: fmt.Sprintf("%d", *reqDTO.Booking.MaxSlotsPerEvent)})
//	}
//	if reqDTO.Booking.AllowOverbooking != nil {
//		updates = append(updates, models.Setting{
//			Category: "booking", Key: "allow_overbooking", Value: fmt.Sprintf("%t", *reqDTO.Booking.AllowOverbooking)})
//	}
//	if reqDTO.Booking.CancellationAllowedHours != nil {
//		updates = append(updates, models.Setting{
//			Category: "booking", Key: "cancellation_allowed_hours", Value: fmt.Sprintf("%d", *reqDTO.Booking.CancellationAllowedHours)})
//	}
//
//	// Проверяем general
//	if reqDTO.General.SiteName != nil {
//		updates = append(updates, models.Setting{Category: "general", Key: "site_name", Value: *reqDTO.General.SiteName})
//	}
//	if reqDTO.General.ContactEmail != nil {
//		updates = append(updates, models.Setting{Category: "general", Key: "contact_email", Value: *reqDTO.General.ContactEmail})
//	}
//	if reqDTO.General.ContactPhone != nil {
//		updates = append(updates, models.Setting{Category: "general", Key: "contact_phone", Value: *reqDTO.General.ContactPhone})
//	}
//
//	return updates
//}
