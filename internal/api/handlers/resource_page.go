package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/resourcepage"
	rp "github.com/yandex-development-1-team/go/internal/resourcepage"
	"github.com/yandex-development-1-team/go/internal/service"
)

type ResourcePageHandler struct {
	service *service.ResourcePageService
}

func NewResourcePageHandler(service *service.ResourcePageService) *ResourcePageHandler {
	return &ResourcePageHandler{service: service}
}

// listResourcePages handles GET /api/v1/resources
func (h *ResourcePageHandler) ListResourcePages(c *gin.Context) {
	ctx := c.Request.Context()
	summaries, err := h.service.GetAllSummaries(ctx)
	if err != nil {
		log.Printf("ERROR: failed to get resource page summaries: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, summaries)
}

// getResourcePage handles GET /api/v1/resources/{slug}
func (h *ResourcePageHandler) GetResourcePage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing slug"})
		return
	}

	ctx := c.Request.Context()
	page, err := h.service.GetResourcePage(ctx, slug)
	if err != nil {
		if fmt.Sprintf("%v", err) == fmt.Sprintf("resource page with slug '%s' not found", slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}
		log.Printf("ERROR: failed to get resource page %s: %v", slug, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	apiResponse := toAPIResponse(page)

	c.JSON(http.StatusOK, apiResponse)
}

// updateResourcePage handles PUT /api/v1/resources/{slug}
func (h *ResourcePageHandler) UpdateResourcePage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing slug"})
		return
	}

	var updateReq resourcepage.UpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	ctx := c.Request.Context()
	newPageData, err := toDomainFromAPIUpdateRequest(&updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse update request: %v", err)})
		return
	}

	updatedPage, err := h.service.UpdateResourcePage(ctx, slug, newPageData)
	if err != nil {
		if fmt.Sprintf("%v", err) == fmt.Sprintf("resource page with slug '%s' not found", slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}
		if strings.Contains(err.Error(), "validation error:") {
			c.JSON(http.StatusBadRequest, gin.H{"field_errors": gin.H{"links": err.Error()}})
			return
		}
		log.Printf("ERROR: failed to update resource page %s: %v", slug, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	apiResponse := toAPIResponse(updatedPage)
	c.JSON(http.StatusOK, apiResponse)
}

// getPublicResourcePage handles GET /api/v1/public/resources/{slug}
func (h *ResourcePageHandler) GetPublicResourcePage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing slug"})
		return
	}

	ctx := c.Request.Context()
	page, err := h.service.GetResourcePage(ctx, slug)
	if err != nil {
		if fmt.Sprintf("%v", err) == fmt.Sprintf("resource page with slug '%s' not found", slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}
		log.Printf("ERROR: failed to get public resource page %s: %v", slug, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	apiResponse := toAPIResponsePublic(page)

	c.JSON(http.StatusOK, apiResponse)
}

// deleteLink handles DELETE /api/v1/resources/{slug}/{id}
func (h *ResourcePageHandler) DeleteLink(c *gin.Context) {
	slug := c.Param("slug")
	id := c.Param("id")

	if slug == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug or Link ID cannot be empty"})
		return
	}

	ctx := c.Request.Context()
	err := h.service.DeleteLink(ctx, slug, id)
	if err != nil {
		if fmt.Sprintf("%v", err) == fmt.Sprintf("resource page with slug '%s' not found", slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page Not Found"})
			return
		}
		if strings.Contains(err.Error(), "not found in page") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Link Not Found"})
			return
		}
		log.Printf("ERROR: failed to delete link %s from page %s: %v", id, slug, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toAPIResponse(domainModel *rp.ResourcePage) *rp.Response {
	if domainModel == nil {
		return nil
	}

	var contentPtr *string
	if domainModel.Content != "" {
		contentPtr = &domainModel.Content
	}

	var linksPtr *[]rp.Link
	if len(domainModel.Links) > 0 {
		linksPtr = &domainModel.Links
	}

	return &rp.Response{
		Title:   domainModel.Title,
		Content: contentPtr,
		Links:   linksPtr,
	}
}

func toAPIResponsePublic(domainModel *rp.ResourcePage) *rp.ResponsePublic {
	if domainModel == nil {
		return nil
	}

	var contentPtr *string
	if domainModel.Content != "" {
		contentPtr = &domainModel.Content
	}

	var linksPtr *[]rp.Link
	if len(domainModel.Links) > 0 {
		linksPtr = &domainModel.Links
	}

	return &rp.ResponsePublic{
		Title:   domainModel.Title,
		Content: contentPtr,
		Links:   linksPtr,
	}
}

func toDomainFromAPIUpdateRequest(apiReq *rp.UpdateRequest) (*rp.ResourcePage, error) {
	if apiReq == nil {
		return nil, nil
	}

	domainUpdate := &rp.ResourcePage{}

	if apiReq.Title != nil {
		domainUpdate.Title = *apiReq.Title
	}
	if apiReq.Content != nil {
		domainUpdate.Content = *apiReq.Content
	}
	if apiReq.Links != nil {

		for _, newLink := range *apiReq.Links {
			parsedURL, err := url.ParseRequestURI(newLink.URL)
			if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
				return nil, fmt.Errorf("validation error: invalid URI '%s' for link ID '%s'", newLink.URL, newLink.ID)
			}
		}

		newLinks := make([]rp.Link, len(*apiReq.Links))
		copy(newLinks, *apiReq.Links)
		domainUpdate.Links = newLinks
	}

	return domainUpdate, nil
}
