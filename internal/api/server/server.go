package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/api/middleware"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/service"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

type APIServices struct {
	BoxService        *apiService.APIBoxService
	SpecialProjectSvc *service.SpecialProjectService
	SettingsService   *apiService.SettingsService
	AnalyticsSvc      *apiService.AnalyticsService
	RecPageSvc        *service.ResourcePageService
	UserSvc           *apiService.UserService
	FileService       *apiService.FileService
}

type Server struct {
	router      *gin.Engine
	services    *APIServices
	srv         *http.Server
	authService *apiService.AuthService
}

func New(cfg *config.Config, services *APIServices, authService *apiService.AuthService) *Server {
	if cfg.Environment == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())
	router.Use(middleware.CORS(cfg.CORS))

	return &Server{
		router:      router,
		services:    services,
		authService: authService,
	}
}

func (s *Server) RegisterRoutes() {
	authHandler := handlers.NewAuthHandler(s.authService)
	boxHandler := handlers.NewBoxHandler(s.services.BoxService)
	specProjHandler := handlers.NewSpecialProjectHandler(s.services.SpecialProjectSvc)
	settingsHandler := handlers.NewSettingsHandler(s.services.SettingsService)
	analyticsHandler := handlers.NewAnalyticsHandler(s.services.AnalyticsSvc)
	recPageHandler := handlers.NewResourcePageHandler(s.services.RecPageSvc)
	userHandler := handlers.NewUserHandler(s.services.UserSvc)

	var fileHandler *handlers.FileHandler
	if s.services.FileService != nil {
		fileHandler = handlers.NewFileHandler(s.services.FileService)
	}

	SetupRoutes(s.router, s.authService.JwtSecret, authHandler, boxHandler, specProjHandler, settingsHandler, analyticsHandler, recPageHandler, userHandler, fileHandler)

}

func (s *Server) Run(cfg *config.Config) error {
	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: s.router,
	}

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
