package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
	"go.uber.org/zap"
)

type SpecialProjectHandler struct {
	svc *service.SpecialProjectService
}

func NewSpecialProjectHandler(svc *service.SpecialProjectService) *SpecialProjectHandler {
	return &SpecialProjectHandler{svc: svc}
}

// PUT /special-projects/{id}
func (h *SpecialProjectHandler) UpdateSpecialProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrInvalidInput.Error()})
		return
	}

	var specialProject models.SpecialProject
	if err := c.ShouldBindJSON(&specialProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrInvalidInput.Error()})
		return
	}
	updatedSpecialProject, err := h.svc.UpdateSpecialProject(c.Request.Context(), id, specialProject)
	if err != nil {
		sendErr(c, id, err)
		return
	}
	c.JSON(http.StatusOK, updatedSpecialProject)
}

// DELETE /special-projects/{id}
func (h *SpecialProjectHandler) DeleteSpecialProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrInvalidInput.Error()})
		return
	}
	if err := h.svc.DeleteSpecialProject(c.Request.Context(), id); err != nil {
		sendErr(c, id, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func sendErr(c *gin.Context, id int, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrInvalidInput.Error()})
	case errors.Is(err, models.ErrSpecialProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": http.StatusText(http.StatusNotFound)})
	default:
		logger.Error("failed to update special project", zap.Error(err), zap.Int("project_id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
	}
}
