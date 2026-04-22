package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/api/middleware"
	"github.com/yandex-development-1-team/go/internal/models"
)

func SetupRoutes(client *sqlx.DB, router *gin.Engine, jwtSecret []byte, authHandler *handlers.AuthHandler, boxHandler *handlers.BoxHandler, specProjHandler *handlers.SpecialProjectHandler, settingsHandler *handlers.SettingsHandler, analyticsHandler *handlers.AnalyticsHandler, recPageHandler *handlers.ResourcePageHandler, userHandler *handlers.UserHandler, fileHandler *handlers.FileHandler, applicationHandler *handlers.ApplicationHandler, bookingHandler *handlers.BookingHandler) {
	middlewareRepo := middleware.NewMiddlewareRepository(client)
	apiV1 := router.Group("/api/v1")
	{
		setupAuthRoutes(apiV1, authHandler)

		protected := apiV1.Group("/")
		protected.Use(middlewareRepo.Auth(jwtSecret))
		{
			setupBoxRoutes(protected, boxHandler, middlewareRepo)
			setupSpecialProjectRoutes(protected, specProjHandler, middlewareRepo)
			setupSettingsRoutes(protected, settingsHandler)
			setupAnalyticsRoutes(protected, analyticsHandler, middlewareRepo)
			setupUserRoutes(protected, userHandler)
			setupResourcesRoutes(protected, recPageHandler)
			setupFileRoutes(protected, fileHandler)
			setupApplicationRoutes(protected, applicationHandler, middlewareRepo)
			setupBookingRoutes(protected, bookingHandler, middlewareRepo)
		}
		public := apiV1.Group("/public")
		public.GET("/resources/:slug", recPageHandler.GetPublicBySlug)
		public.POST("/applications/", applicationHandler.Create)
	}
}

func setupAuthRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.HandleLogin)
		auth.POST("/register", h.RegisterHandler)
		auth.POST("/refresh", h.HandleRefresh)
		auth.POST("/logout", h.HandleLogout)
		auth.POST("/forgot-password", h.HandleForgotPassword)
		auth.POST("/reset-password", h.HandleResetPassword)
	}
}

func setupSpecialProjectRoutes(rg *gin.RouterGroup, h *handlers.SpecialProjectHandler, middlewareRepo *middleware.Middleware) {
	sp := rg.Group("/special-projects")
	{
		sp.GET("/", middlewareRepo.RoleVerification(models.PermSpecProjectView), h.ListSpecialProjects)
		sp.POST("/", middlewareRepo.RoleVerification(models.PermSpecProjectEdit), h.CreateSpecialProject)
		sp.GET("/:id", middlewareRepo.RoleVerification(models.PermSpecProjectView), h.GetSpecialProjectByID)
		sp.PUT("/:id", middlewareRepo.RoleVerification(models.PermSpecProjectEdit), h.UpdateSpecialProject)
		sp.DELETE("/:id", middlewareRepo.RoleVerification(models.PermSpecProjectDelete), h.DeleteSpecialProject)
	}
}

func setupAnalyticsRoutes(rg *gin.RouterGroup, h *handlers.AnalyticsHandler, middlewareRepo *middleware.Middleware) {
	analytics := rg.Group("/analytics")
	{
		analytics.GET("/export", middlewareRepo.RoleVerification(models.PermAnalyticsView), h.Export)
	}
}

func setupBoxRoutes(rg *gin.RouterGroup, boxHandler *handlers.BoxHandler, middlewareRepo *middleware.Middleware) {
	boxes := rg.Group("/boxes")
	{
		boxes.GET("/", middleware.RequireManagersOrAdmin(), boxHandler.List)
		boxes.POST("/", middlewareRepo.RoleVerification(models.PermBoxesCreate), boxHandler.Create)
		boxes.GET("/:id", middleware.RequireManagersOrAdmin(), boxHandler.GetByID)
		boxes.PUT("/:id", middlewareRepo.RoleVerification(models.PermBoxesEdit), boxHandler.Update)
		boxes.DELETE("/:id", middlewareRepo.RoleVerification(models.PermBoxesDelete), boxHandler.Delete)
		boxes.POST("/:id/image", middlewareRepo.RoleVerification(models.PermBoxesEdit), boxHandler.UploadImage)
		boxes.PUT("/:id/status", middlewareRepo.RoleVerification(models.PermBoxesEdit), boxHandler.UpdateStatus)
	}
}

func setupSettingsRoutes(rg *gin.RouterGroup, settingsHandler *handlers.SettingsHandler) {
	settings := rg.Group("/settings")
	{
		settings.GET("/", middleware.RequireManagersOrAdmin(), settingsHandler.Get)
		settings.PUT("/", middleware.RequireAdmin(), settingsHandler.Put)
		settings.POST("/", middleware.RequireAdmin(), settingsHandler.Post)
	}
}

func setupUserRoutes(rg *gin.RouterGroup, h *handlers.UserHandler) {
	users := rg.Group("/users")
	{
		users.GET("/", middleware.RequireAdmin(), h.List)
		users.GET("/:id", middleware.RequireAdmin(), h.GetByID)
	}
}

func setupResourcesRoutes(rg *gin.RouterGroup, h *handlers.ResourcePageHandler) {
	resources := rg.Group("/resources")
	{
		resources.GET("/", middleware.RequireManagersOrAdmin(), h.GetAll)
		resources.GET("/:slug", middleware.RequireManagersOrAdmin(), h.GetBySlug)
		resources.PUT("/:slug", middleware.RequireManagersOrAdmin(), h.Update)
		resources.PUT("/:slug/file", middleware.RequireManagersOrAdmin(), h.UploadFile)
		resources.DELETE("/:slug/:id", middleware.RequireManagersOrAdmin(), h.DeleteLink)
		resources.DELETE("/:slug", middleware.RequireManagersOrAdmin(), h.Delete)
	}
}

func setupFileRoutes(rg *gin.RouterGroup, h *handlers.FileHandler) {
	files := rg.Group("/files")
	{
		files.POST("/upload", middleware.RequireManagersOrAdmin(), h.Upload)
	}
}

func setupApplicationRoutes(rg *gin.RouterGroup, h *handlers.ApplicationHandler, middlewareRepo *middleware.Middleware) {
	applications := rg.Group("/applications")
	{
		applications.GET("/", middlewareRepo.RoleVerification(models.PermSpecProjectView), h.ApplicationsList)
		applications.POST("/", middlewareRepo.RoleVerification(models.PermSpecProjectEdit), h.Create)
		applications.GET("/:id", middlewareRepo.RoleVerification(models.PermSpecProjectView), h.GetByID)
		applications.PUT("/:id/status", middlewareRepo.RoleVerification(models.PermSpecProjectEdit), h.UpdateApplicationStatus)
		applications.DELETE("/:id", middlewareRepo.RoleVerification(models.PermSpecProjectDelete), h.DeleteApplication)

	}
}

func setupBookingRoutes(rg *gin.RouterGroup, h *handlers.BookingHandler, middlewareRepo *middleware.Middleware) {
	bookings := rg.Group("/bookings")
	{
		bookings.GET("/", middlewareRepo.RoleVerification(models.PermBookingsView), h.BookingsList)
		bookings.GET("/:id", middlewareRepo.RoleVerification(models.PermBookingsView), h.BookingsById)
		bookings.PUT("/:id/status", middlewareRepo.RoleVerification(models.PermBookingsEdit), h.UpdateBookingStatus)
		bookings.DELETE("/:id", middlewareRepo.RoleVerification(models.PermBookingsDelete), h.DeleteBooking)
	}
}
