package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type ApplicationHandler struct {
	svc   *svcapi.ApplicationsService
	token string
}

func NewApplicationHandler(svc *svcapi.ApplicationsService, token string) *ApplicationHandler {
	return &ApplicationHandler{
		svc:   svc,
		token: token,
	}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	token := c.GetHeader("X-Webhook-Token")
	formAnswerID := c.GetHeader("X-Form-Answer-Id")
	if formAnswerID == "" {
		logger.Warn("create application: missing form answer id")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		c.Abort()
		return
	}

	if token != h.token {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		c.Abort()
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Warn("failed to read body")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}
	cleaned := strings.ReplaceAll(string(body), `\"`, `"`)

	var payload dto.CreateApplicationRequest
	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		logger.Warn("create application bad request", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	app := &models.Application{
		FormAnswerId: formAnswerID,
		CustomerName: payload.Answer.Data.FirstName.Value + " " + payload.Answer.Data.LastName.Value,
		ContactInfo:  payload.Answer.Data.Telegram.Value,
		Description:  payload.Answer.Data.Description.Value,
	}

	err = h.svc.Create(c.Request.Context(), app)
	if err != nil {
		logger.Warn("create application error",
			zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ApplicationHandler) GetByID(c *gin.Context) {
	var uri models.ApplicationURI
	if err := c.ShouldBindUri(&uri); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	app, err := h.svc.GetApplicationByID(c.Request.Context(), uri.ID)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toAppResponse(app))
}

func (h *ApplicationHandler) UpdateApplicationStatus(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	var status dto.ApplicationUpdateStatus
	if err = c.ShouldBindJSON(&status); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	app, err := h.svc.UpdateApplicationStatus(c.Request.Context(), id, status.Status)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toAppResponse(app))
}

func (h *ApplicationHandler) ApplicationsList(c *gin.Context) {
	var query dto.ApplicationListRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	list, err := h.svc.ApplicationsList(c.Request.Context(), toAppFilter(&query))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toApplicationListResponse(list))
}

func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	var uri models.ApplicationURI
	if err := c.ShouldBindUri(&uri); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	err := h.svc.DeleteApplication(c.Request.Context(), uri.ID)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func toApplicationListResponse(result *models.ApplicationList) dto.ApplicationListResponse {
	items := make([]dto.ApplicationListItem, len(result.Items))
	for i, app := range result.Items {
		items[i] = dto.ApplicationListItem{
			ID:           app.ID,
			Status:       app.Status,
			ManagerID:    app.ManagerID,
			ManagerName:  app.ManagerName,
			CustomerName: app.CustomerName,
			ContactInfo:  app.ContactInfo,
			CreatedAt:    app.CreatedAt,
		}
	}

	return dto.ApplicationListResponse{
		Items: items,
		Pagination: dto.Pagination{
			Total:  result.Total,
			Limit:  result.Limit,
			Offset: result.Offset,
		},
	}
}

func toAppResponse(app *models.Application) *dto.Application {
	return &dto.Application{
		ID:           app.ID,
		Status:       app.Status,
		ManagerID:    app.ManagerID,
		ManagerName:  app.ManagerName,
		FormAnswerId: app.FormAnswerId,
		CustomerName: app.CustomerName,
		ContactInfo:  app.ContactInfo,
		Description:  app.Description,
		CreatedAt:    app.CreatedAt,
		UpdatedAt:    app.UpdatedAt,
	}
}

func toAppFilter(req *dto.ApplicationListRequest) *models.ApplicationFilter {
	filter := &models.ApplicationFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	if req.Status != nil {
		filter.Status = *req.Status
	}
	if req.ManagerID != nil {
		filter.ManagerID = *req.ManagerID
	}
	if req.CustomerName != nil {
		filter.CustomerName = *req.CustomerName
	}

	return filter
}
