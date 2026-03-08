package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/api/middleware"
	"github.com/yandex-development-1-team/go/internal/config"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

// APIServices contains API services
type APIServices struct {
	// Сервисов будет много, поэтому и обернул в структуру, иначе слишком много параметров получится
	BoxService      *apiService.APIBoxService
	SettingsService *apiService.SettingsService
}

// Server server structure
type Server struct {
	router   *gin.Engine
	services *APIServices
	srv      *http.Server
}

// New creates a new server
func New(env string, services *APIServices) *Server {
	if env == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())
	router.Use(middleware.CORS())

	return &Server{
		router:   router,
		services: services,
	}
}

// RegisterRoutes registers routes
func (s *Server) RegisterRoutes() {
	boxHandler := handlers.NewBoxHandler(s.services.BoxService)
	settingsHandler := handlers.NewSettingsHandler(s.services.SettingsService)

	SetupRoutes(s.router, boxHandler, settingsHandler)
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

// Shutdown stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
