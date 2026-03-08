package handlers

import (
	"github.com/gin-gonic/gin"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

// BoxHandler handles HTTP requests for boxed solutions
type BoxHandler struct {
	boxService *apiService.APIBoxService
}

// NewBoxHandler creates a new 'BoxHandler'
func NewBoxHandler(boxService *apiService.APIBoxService) *BoxHandler {
	return &BoxHandler{
		boxService: boxService,
	}
}

// List GET /api/v1/boxes
func (h *BoxHandler) List(c *gin.Context) {
	// TODO: Implement with filters: status, search, limit, offset, sort, order
	c.JSON(200, gin.H{"message": "BoxHandler.List - not implemented yet"})
}

// Create POST /api/v1/boxes
func (h *BoxHandler) Create(c *gin.Context) {
	// TODO: Implement
	c.JSON(201, gin.H{"message": "BoxHandler.Create - not implemented yet"})
}

// GetByID GET /api/v1/boxes/:id
func (h *BoxHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.GetByID - not implemented yet", "id": id})
}

// Update PUT /api/v1/boxes/:id
func (h *BoxHandler) Update(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.Update - not implemented yet", "id": id})
}

// Delete DELETE /api/v1/boxes/:id
func (h *BoxHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.Delete - not implemented yet", "id": id})
}

// UploadImage POST /api/v1/boxes/:id/image
func (h *BoxHandler) UploadImage(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.UploadImage - not implemented yet", "id": id})
}

// UpdateStatus PUT /api/v1/boxes/:id/status
func (h *BoxHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"message": "BoxHandler.UpdateStatus - not implemented yet", "id": id})
}

// Export GET /api/v1/boxes/export
func (h *BoxHandler) Export(c *gin.Context) {
	// TODO: Handle format param (pdf/csv)
	c.JSON(200, gin.H{"message": "BoxHandler.Export - not implemented yet"})
}
