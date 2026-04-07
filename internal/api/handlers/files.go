package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

// FileHandler handles HTTP requests related to file upload.
type FileHandler struct {
	fileService *apiService.FileService
}

// NewFileHandler creates a new FileHandler instance.
func NewFileHandler(fileService *apiService.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) Upload(c *gin.Context) {
	formFile, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	src, err := formFile.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer func() { _ = src.Close() }()

	contentType := formFile.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	resp, err := h.fileService.Upload(
		c.Request.Context(),
		src,
		formFile.Filename,
		contentType,
		formFile.Size,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
