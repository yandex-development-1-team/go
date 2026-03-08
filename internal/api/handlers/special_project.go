package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
)

type SpecialProjectHandler struct {
	svc *service.SpecialProjectService
}

func NewSpecialProjectHandler(svc *service.SpecialProjectService) *SpecialProjectHandler {
	return &SpecialProjectHandler{svc: svc}
}

// Конвертеры
func toDomain(dto *dto.SpecialProjectCreateRequest) *models.SpecialProject {
	if dto == nil {
		return nil
	}
	return &models.SpecialProject{
		Title:         dto.Title,
		Description:   dto.Description,
		Image:         dto.Image,
		IsActiveInBot: dto.IsActiveInBot,
	}
}

func fromDomain(domain *models.SpecialProject) *dto.SpecialProjectResponse {
	if domain == nil {
		return nil
	}

	return &dto.SpecialProjectResponse{
		ID:            domain.ID,
		Title:         domain.Title,
		Description:   domain.Description,
		Image:         domain.Image,
		IsActiveInBot: domain.IsActiveInBot,
	}
}

func fromDomainList(domain []models.SpecialProject) []dto.SpecialProjectListItem {
	if domain == nil {
		return nil
	}

	result := make([]dto.SpecialProjectListItem, 0, len(domain))
	for _, item := range domain {
		result = append(result, dto.SpecialProjectListItem{
			ID:            item.ID,
			Title:         item.Title,
			IsActiveInBot: item.IsActiveInBot,
		})
	}
	return result
}

// POST /api/v1/special-projects
func (h *SpecialProjectHandler) CreateSpecialProject(c *gin.Context) {
	var req dto.SpecialProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	proj := toDomain(&req)
	project, err := h.svc.Create(c.Request.Context(), proj)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	resp := fromDomain(project)
	c.JSON(http.StatusCreated, resp)
}

// GET /api/v1/special-projects
func (h *SpecialProjectHandler) ListSpecialProjects(c *gin.Context) {
	status := c.Query("status") // active | inactive
	search := c.Query("search")

	if status != "" && status != "active" && status != "inactive" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Недопустимый статус"})
		return
	}

	domainlist, err := h.svc.List(c.Request.Context(), status, search)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	list := fromDomainList(domainlist)

	c.JSON(http.StatusOK, list)
}

// GET /api/v1/special-projects/{id}
func (h *SpecialProjectHandler) GetSpecialProjectByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор"})
		return
	}

	project, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	resp := fromDomain(project)
	c.JSON(http.StatusOK, resp)
}
