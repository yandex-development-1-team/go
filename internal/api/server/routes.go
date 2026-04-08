package server

import (
	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/api/middleware"
)

func SetupRoutes(router *gin.Engine, jwtSecret []byte, authHandler *handlers.AuthHandler, boxHandler *handlers.BoxHandler, specProjHandler *handlers.SpecialProjectHandler, settingsHandler *handlers.SettingsHandler, analyticsHandler *handlers.AnalyticsHandler, recPageHandler *handlers.ResourcePageHandler, userHandler *handlers.UserHandler, fileHandler *handlers.FileHandler, applicationHandler *handlers.ApplicationHandler) {
	apiV1 := router.Group("/api/v1")
	{
		setupAuthRoutes(apiV1, authHandler)

		protected := apiV1.Group("/")
		protected.Use(middleware.Auth(jwtSecret))
		{
			setupBoxRoutes(protected, boxHandler)
			setupSpecialProjectRoutes(protected, specProjHandler)
			setupSettingsRoutes(protected, settingsHandler)
			setupAnalyticsRoutes(protected, analyticsHandler)
			setupUserRoutes(protected, userHandler)
			setupResourcesRoutes(protected, recPageHandler)
			setupFileRoutes(protected, fileHandler)
			setupApplicationRoutes(protected, applicationHandler)
		}
		public := apiV1.Group("/public")
		public.GET("/resources/:slug", recPageHandler.GetPublicResourcePage)
	}
}

func setupAuthRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.HandleLogin)
		auth.POST("/register", h.RegisterHandler)
		auth.POST("/refresh", h.HandleRefresh)
		auth.POST("/logout", h.HandleLogout)
	}
}

func setupSpecialProjectRoutes(rg *gin.RouterGroup, h *handlers.SpecialProjectHandler) {
	sp := rg.Group("/special-projects")
	{
		sp.GET("/", h.ListSpecialProjects)
		sp.POST("/", h.CreateSpecialProject)
		sp.GET("/:id", h.GetSpecialProjectByID)
	}
}

func setupAnalyticsRoutes(rg *gin.RouterGroup, h *handlers.AnalyticsHandler) {
	analytics := rg.Group("/analytics")
	{
		analytics.GET("/export", h.Export)
	}
}

func setupBoxRoutes(rg *gin.RouterGroup, boxHandler *handlers.BoxHandler) {
	boxes := rg.Group("/boxes")
	{
		boxes.GET("/", boxHandler.List)
		boxes.POST("/")
		boxes.GET("/:id", boxHandler.GetByID)
		boxes.PUT("/:id", boxHandler.Update)
		boxes.DELETE("/:id", boxHandler.Delete)
		boxes.POST("/:id/image", boxHandler.UploadImage)
		boxes.PUT("/:id/status", boxHandler.UpdateStatus)
	}
}

func setupSettingsRoutes(rg *gin.RouterGroup, settingsHandler *handlers.SettingsHandler) {
	settings := rg.Group("/settings")
	{
		settings.GET("/", settingsHandler.Get)
		settings.PUT("/", middleware.RequireAdmin(), settingsHandler.Put)
	}
}

func setupUserRoutes(rg *gin.RouterGroup, h *handlers.UserHandler) {
	users := rg.Group("/users")
	{
		users.GET("/", h.List)
		users.GET("/:id", h.GetByID)
	}
}

func setupResourcesRoutes(rg *gin.RouterGroup, h *handlers.ResourcePageHandler) {
	resources := rg.Group("/resources")
	{
		resources.GET("/", h.ListResourcePages)
		resources.GET("/:slug", h.GetResourcePage)
		resources.PUT("/:slug", h.UpdateResourcePage)
		resources.DELETE("/:slug", h.DeleteLink)
	}
}

func setupFileRoutes(rg *gin.RouterGroup, h *handlers.FileHandler) {
	files := rg.Group("/files")
	{
		files.POST("/upload", h.Upload)
	}
}

func setupApplicationRoutes(rg *gin.RouterGroup, h *handlers.ApplicationHandler) {
	applications := rg.Group("/applications")
	{
		applications.GET("/", h.List)
		applications.POST("/", h.Create)
		applications.GET("/:id", h.GetByID)
		applications.PUT("/:id", h.Update)
		applications.DELETE("/:id", h.Delete)
	}
}
