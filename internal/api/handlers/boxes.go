package handlers

import (
	"net/http"
	"strconv"

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
	list, err := h.boxService.List(c.Request.Context())
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *BoxHandler) Create(c *gin.Context) {
	c.JSON(201, gin.H{"message": "BoxHandler.Create - not implemented yet"})
}

func (h *BoxHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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

	c.JSON(http.StatusOK, box)
}

func (h *BoxHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	var req dto.BoxUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	box, err := h.boxService.Update(c.Request.Context(), id, &req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	if box == nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusNotFound, []string{"Коробочное решение не найдено"})
		return
	}

	c.JSON(http.StatusOK, box)
}

func (h *BoxHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверный формат запроса"})
		return
	}

	if req.Status != "active" && req.Status != "hidden" && req.Status != "draft" && req.Status != "processed" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный статус. Допустимые значения: active, hidden, draft, processed"})
		return
	}

	result, err := h.boxService.UpdateStatus(c.Request.Context(), id, models.ServiceStatus(req.Status))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	if result == nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusNotFound, []string{"Коробочное решение не найдено"})
		return
	}

	c.JSON(http.StatusOK, result)
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
