package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

type BoxHandler struct {
	boxService *apiService.APIBoxService
}

func NewBoxHandler(boxService *apiService.APIBoxService) *BoxHandler {
	return &BoxHandler{boxService: boxService}
}

func (h *BoxHandler) List(c *gin.Context) {
	var req dto.BoxListQuery

	if err := c.ShouldBindQuery(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	query := models.BoxList{
		Status: req.Status,
		Search: req.Search,
		Limit:  req.Limit,
		Offset: req.Offset,
		Sort:   req.Sort,
		Order:  req.Order,
	}

	result, err := h.boxService.List(c.Request.Context(), query)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	response := &dto.BoxListResponse{
		Items: result.Items,
		Pagination: dto.Pagination{
			Limit:  result.Limit,
			Offset: result.Offset,
			Total:  result.Total,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *BoxHandler) Create(c *gin.Context) {
	var req dto.BoxCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	service, err := h.boxService.Create(c.Request.Context(), toServiceCreateModel(&req))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusOK, toBoxResponse(service))
}

func (h *BoxHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор коробочного решения"})
		return
	}

	box, err := h.boxService.GetByID(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	if box == nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusNotFound, []string{"Коробочное решение не найдено"})
		return
	}

	c.JSON(http.StatusOK, toBoxResponse(box))
}

func (h *BoxHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	var req dto.BoxUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	box, err := h.boxService.Update(c.Request.Context(), id, toServiceUpdateModel(&req))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	// if box == nil {
	// 	apierrors.WriteErrorMessagesGin(c, http.StatusNotFound, []string{"Коробочное решение не найдено"})
	// 	return
	// }

	c.JSON(http.StatusOK, toBoxResponse(box))
}

func (h *BoxHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	err = h.boxService.Delete(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Коробочное решение успешно удалено"})
}

func (h *BoxHandler) UploadImage(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.UploadImage - not implemented yet", "id": id})
}

func (h *BoxHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	var req dto.BoxUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	result, err := h.boxService.UpdateStatus(c.Request.Context(), id, models.ServiceStatus(req.Status))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, &dto.BoxStatusResponse{
		ID:        result.ID,
		Status:    result.Status,
		UpdatedAt: result.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *BoxHandler) Export(c *gin.Context) {
	status := c.Query("status")
	format := c.DefaultQuery("format", "pdf")

	if format != "pdf" && format != "csv" {
		format = "pdf"
	}

	if status != "" {
		if status != "active" && status != "hidden" && status != "draft" && status != "processed" {
			apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный статус"})
			return
		}
	}

	data, contentType, err := h.boxService.Export(c.Request.Context(), status, format)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=boxes."+format)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

func toBoxResponse(box *models.Service) *dto.BoxDetailResponse {
	var slots []dto.BoxAvailableSlot
	for _, s := range box.BoxAvailableSlots {
		slots = append(slots, dto.BoxAvailableSlot{
			Date:      s.Date,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		})
	}

	return &dto.BoxDetailResponse{
		ID:                box.ID,
		Name:              box.Name,
		Description:       box.Description,
		Rules:             box.Rules,
		BoxAvailableSlots: slots,
		Location:          box.Location,
		Price:             box.Price,
		Image:             box.Image,
		Status:            box.Status,
		Organizer:         box.Organizer,
		CreatedAt:         box.CreatedAt,
		UpdatedAt:         box.UpdatedAt,
	}
}

func toServiceUpdateModel(box *dto.BoxUpdateRequest) *models.BoxUpdate {
	var slots []models.BoxAvailableSlot
	for _, s := range box.Slots {
		slots = append(slots, models.BoxAvailableSlot{
			Date:      s.Date,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		})
	}

	return &models.BoxUpdate{
		Name:        box.Name,
		Description: box.Description,
		Rules:       box.Rules,
		Slots:       slots,
		Location:    box.Location,
		Price:       box.Price,
		Image:       box.Image,
		Status:      (*string)(box.Status),
		Organizer:   box.Organizer,
	}
}

func toServiceCreateModel(box *dto.BoxCreateRequest) *models.BoxCreate {
	var slots []models.BoxAvailableSlot
	for _, s := range box.Slots {
		slots = append(slots, models.BoxAvailableSlot{
			Date:      s.Date,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		})
	}

	return &models.BoxCreate{
		Name:        box.Name,
		Slug:        StringPtr("slug"),
		Description: box.Description,
		Rules:       box.Rules,
		Slots:       slots,
		Location:    box.Location,
		Price:       box.Price,
		Image:       box.Image,
		Status:      box.Status,
		Organizer:   box.Organizer,
	}
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}
