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

// APIServices contains API services
type APIServices struct {
	BoxService        *apiService.APIBoxService
	SpecialProjectSvc *service.SpecialProjectService
	SettingsService   *apiService.SettingsService
}

// Server server structure
type Server struct {
	router      *gin.Engine
	services    *APIServices
	srv         *http.Server
	authService *apiService.AuthService
}

// New creates a new server (CORS и прочие настройки берутся из cfg).
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

// RegisterRoutes registers routes
func (s *Server) RegisterRoutes() {
	boxHandler := handlers.NewBoxHandler(s.services.BoxService)
	specProjHandler := handlers.NewSpecialProjectHandler(s.services.SpecialProjectSvc)
	settingsHandler := handlers.NewSettingsHandler(s.services.SettingsService)

	SetupRoutes(s.router, s.authService.JwtSecret, boxHandler, specProjHandler, settingsHandler)
}

// Run starts the server
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

// Shutdown stops the server. No-op if Run was not called.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
