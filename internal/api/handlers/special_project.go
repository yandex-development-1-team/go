package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
	"github.com/yandex-development-1-team/go/internal/specialproject"
)

type SpecialProjectHandler struct {
	svc *service.SpecialProjectService
}

func NewSpecialProjectHandler(svc *service.SpecialProjectService) *SpecialProjectHandler {
	return &SpecialProjectHandler{svc: svc}
}

func toDomain(req *specialproject.CreateRequest) *specialproject.Project {
	if req == nil {
		return nil
	}
	return &specialproject.Project{
		Title:         req.Title,
		Description:   req.Description,
		Image:         req.Image,
		IsActiveInBot: req.IsActiveInBot,
	}
}

func toResponse(domain *specialproject.Project) *specialproject.Response {
	if domain == nil {
		return nil
	}
	return &specialproject.Response{
		ID:            domain.ID,
		Title:         domain.Title,
		Description:   domain.Description,
		Image:         domain.Image,
		IsActiveInBot: domain.IsActiveInBot,
		CreatedAt:     domain.CreatedAt,
		UpdatedAt:     domain.UpdatedAt,
	}
}

func toItemList(domain []*specialproject.Project) []specialproject.ListItem {
	if domain == nil {
		return nil
	}
	result := make([]specialproject.ListItem, 0, len(domain))
	for _, item := range domain {
		result = append(result, specialproject.ListItem{
			ID:            item.ID,
			Title:         item.Title,
			IsActiveInBot: item.IsActiveInBot,
		})
	}
	return result
}

// POST /api/v1/special-projects
func (h *SpecialProjectHandler) CreateSpecialProject(c *gin.Context) {
	var req specialproject.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": err.Error()})
		return
	}
	proj := toDomain(&req)
	project, err := h.svc.Create(c.Request.Context(), proj)
	if err != nil {
		sendSpecialProjectErr(c, &proj.ID, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(project))
}

// GET /api/v1/special-projects
func (h *SpecialProjectHandler) ListSpecialProjects(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	if status != "" && status != "active" && status != "inactive" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status"})
		return
	}

	limitP := c.Query("limit")
	limit, err := strconv.Atoi(limitP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
		return
	}

	offsetP := c.Query("offset")
	offset, err := strconv.Atoi(offsetP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
		return
	}

	if limit < 0 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status"})
		return
	}

	domainlist, total, err := h.svc.List(c.Request.Context(), status, search, limit, offset)
	if err != nil {
		sendSpecialProjectErr(c, nil, err)
		return
	}
	items := toItemList(domainlist)

	c.JSON(http.StatusOK, specialproject.ResponseList{Items: items,
		Pagination: specialproject.Pagination{Total: total, Limit: limit, Offset: offset}})
}

// GET /api/v1/special-projects/{id}
func (h *SpecialProjectHandler) GetSpecialProjectByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	project, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		sendSpecialProjectErr(c, &id, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(project))
}

// PUT /api/v1/special-projects/{id}
func (h *SpecialProjectHandler) UpdateSpecialProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	var payload specialproject.Project
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": err.Error()})
		return
	}
	updated, err := h.svc.Update(c.Request.Context(), id, &payload)
	if err != nil {
		sendSpecialProjectErr(c, &id, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(updated))
}

// DELETE /api/v1/special-projects/{id}
func (h *SpecialProjectHandler) DeleteSpecialProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		sendSpecialProjectErr(c, &id, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func sendSpecialProjectErr(c *gin.Context, id *int64, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrInvalidInput.Error()})
	case errors.Is(err, specialproject.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": http.StatusText(http.StatusNotFound)})
	default:
		if id != nil {
			logger.Error("special project handler error", zap.Error(err), zap.Int64("project_id", *id))
		} else {
			logger.Error("special project handler error", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
	}
}
