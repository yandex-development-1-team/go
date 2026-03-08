package server

import (
	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/api/middleware"
)

// SetupRoutes configures all API routes according to openapi.json
func SetupRoutes(router *gin.Engine, boxHandler *handlers.BoxHandler) {
	// API v1 group
	apiV1 := router.Group("/api/v1")
	{
		// Protected routes
		protected := apiV1.Group("/")
		protected.Use(middleware.Auth())
		{
			setupBoxRoutes(protected, boxHandler)
		}
	}
}

// setupBoxRoutes routes for boxed solutions
func setupBoxRoutes(rg *gin.RouterGroup, boxHandler *handlers.BoxHandler) {
	boxes := rg.Group("/boxes")
	{
		// GET /api/v1/boxes - список коробок с фильтрацией
		boxes.GET("/", boxHandler.List)

		// POST /api/v1/boxes - создание коробки
		boxes.POST("/")

		// GET /api/v1/boxes/export - экспорт коробок
		boxes.GET("/export", boxHandler.Export)

		// Маршруты с параметром :id
		boxes.GET("/:id", boxHandler.GetByID)             // GET /api/v1/boxes/{id}
		boxes.PUT("/:id", boxHandler.Update)              // PUT /api/v1/boxes/{id}
		boxes.DELETE("/:id", boxHandler.Delete)           // DELETE /api/v1/boxes/{id}
		boxes.POST("/:id/image", boxHandler.UploadImage)  // POST /api/v1/boxes/{id}/image
		boxes.PUT("/:id/status", boxHandler.UpdateStatus) // PUT /api/v1/boxes/{id}/status
	}
}
