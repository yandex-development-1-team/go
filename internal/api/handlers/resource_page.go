package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
)

type ResourcePageHandler struct {
	service *service.ResourcePageService
}

func NewResourcePageHandler(service *service.ResourcePageService) *ResourcePageHandler {
	return &ResourcePageHandler{service: service}
}

func (h *ResourcePageHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	pages, err := h.service.GetAllResourcePages(ctx)
	if err != nil {
		logger.Error("GetAll", zap.Error(err))
		apierrors.WriteErrorMessagesGin(c, http.StatusInternalServerError, []string{"Internal Server Error"})
		return
	}

	response := make([]dto.ResourcePageResponse, 0, len(pages))
	for _, p := range pages {
		response = append(response, toResourcePageResponse(p))
	}

	c.JSON(http.StatusOK, response)
}

func (h *ResourcePageHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	ctx := c.Request.Context()
	page, err := h.service.GetResourcePage(ctx, slug)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toResourcePageResponse(*page))
}

func (h *ResourcePageHandler) GetPublicBySlug(c *gin.Context) {
	slug := c.Param("slug")

	ctx := c.Request.Context()
	page, err := h.service.GetResourcePage(ctx, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, toResourcePagePublicResponse(*page))
}

func (h *ResourcePageHandler) Update(c *gin.Context) {
	slug := c.Param("slug")

	var req dto.ResourcePageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	ctx := c.Request.Context()
	updatedPage, err := h.service.UpdateResourcePage(ctx, slug, toDomainUpdateRequest(req))
	if err != nil {
		logger.Error("Update", zap.Error(err))
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toResourcePageResponse(*updatedPage))
}

func (h *ResourcePageHandler) DeleteLink(c *gin.Context) {
	slug := c.Param("slug")
	id := c.Param("id")

	ctx := c.Request.Context()
	page, err := h.service.DeleteResourcePageLink(ctx, slug, id)
	if err != nil {
		log.Printf("ERROR: failed to delete link %s from %s: %v", id, slug, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, toResourcePageResponse(*page))
}

func (h *ResourcePageHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")

	ctx := c.Request.Context()
	updatedPage, err := h.service.ClearResourcePage(ctx, slug)
	logger.Error("Delete", zap.Error(err))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toResourcePageResponse(*updatedPage))
}

func toResourcePageResponse(p models.ResourcePage) dto.ResourcePageResponse {
	links := make([]dto.ResourcePageLink, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, dto.ResourcePageLink{
			ID:    l.ID,
			Title: l.Title,
			URL:   l.URL,
		})
	}
	return dto.ResourcePageResponse{
		Slug:      p.Slug,
		Title:     p.Title,
		Content:   p.Content,
		Links:     links,
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func toDomainUpdateRequest(req dto.ResourcePageUpdateRequest) models.ResourcePage {
	links := make([]models.ResourcePageLink, 0, len(req.Links))
	for _, l := range req.Links {
		links = append(links, models.ResourcePageLink{
			ID:    l.ID,
			Title: l.Title,
			URL:   l.URL,
		})
	}
	return models.ResourcePage{
		Title:   req.Title,
		Content: req.Content,
		Links:   links,
	}
}

func toResourcePagePublicResponse(p models.ResourcePage) dto.ResourcePagePublicResponse {
	links := make([]dto.ResourcePageLink, 0, len(p.Links))
	for _, l := range p.Links {
		links = append(links, dto.ResourcePageLink{
			ID:    l.ID,
			Title: l.Title,
			URL:   l.URL,
		})
	}
	return dto.ResourcePagePublicResponse{
		Title:   p.Title,
		Content: p.Content,
		Links:   links,
	}
}
