package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type ApplicationHandler struct {
	repo repository.ApplicationRepository
}

func NewApplicationHandler(repo repository.ApplicationRepository) *ApplicationHandler {
	return &ApplicationHandler{repo: repo}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	var req models.ApplicationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	if !req.Type.Valid() || !req.Source.Valid() || req.CustomerName == "" || req.ContactInfo == "" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные заявки"})
		return
	}

	app, err := h.repo.CreateApplication(c.Request.Context(), &req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusCreated, app)
}

func (h *ApplicationHandler) List(c *gin.Context) {
	if c.Query("date_from") != "" || c.Query("date_to") != "" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	var req dto.ApplicationListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	limit, ok := parseLimit(c.Query("limit"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	offset, ok := parseOffset(c.Query("offset"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	managerID, ok := parseOptionalPositiveInt64(c.Query("manager_id"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	filter := models.ApplicationFilter{
		CustomerName: req.CustomerName,
		Limit:        limit,
		Offset:       offset,
		ManagerID:    managerID,
	}

	if req.Status != nil {
		status := models.ApplicationStatus(*req.Status)
		filter.Status = &status
	}

	if req.Type != nil {
		t := models.ApplicationType(*req.Type)
		filter.Type = &t
	}

	items, total, err := h.repo.GetApplications(c.Request.Context(), filter)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	if items == nil {
		items = []models.Application{}
	}

	c.JSON(http.StatusOK, dto.ApplicationListResponse{
		Items: items,
		Pagination: dto.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *ApplicationHandler) GetByID(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор заявки"})
		return
	}

	app, err := h.repo.GetApplicationByID(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор заявки"})
		return
	}

	var req models.ApplicationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}
	if !req.HasUpdates() {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Нужно передать хотя бы одно поле для обновления"})
		return
	}
	if req.Status != nil && !req.Status.Valid() {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный статус заявки"})
		return
	}

	app, err := h.repo.UpdateApplication(c.Request.Context(), id, &req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор заявки"})
		return
	}

	if err := h.repo.DeleteApplication(c.Request.Context(), id); err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}

func parsePositiveID(raw string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parseLimit(raw string) (int, bool) {
	if raw == "" {
		return dto.DefaultApplicationLimit, true
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if limit < 1 || limit > dto.MaxApplicationLimit {
		return 0, false
	}

	return limit, true
}

func parseOffset(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}

	offset, err := strconv.Atoi(raw)
	if err != nil || offset < 0 {
		return 0, false
	}

	return offset, true
}

func parseOptionalPositiveInt64(raw string) (*int64, bool) {
	if raw == "" {
		return nil, true
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}

	return &id, true
}
