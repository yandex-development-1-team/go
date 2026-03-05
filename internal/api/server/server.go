package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"github.com/yandex-development-1-team/go/internal/service"
)

// Config server config
type Config struct {
	Port       string
	Env        string
	BoxService *service.APIBoxService
}

// Server server structure
type Server struct {
	router *gin.Engine
	cfg    Config
	srv    *http.Server
}

// New creates a new server
func New(cfg Config) *Server {
	if cfg.Env == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	// router.Use(middleware.Logger())
	// router.Use(middleware.Metrics())
	// router.Use(middleware.CORS())

	return &Server{
		router: router,
		cfg:    cfg,
	}
}

// RegisterRoutes registers routes
func (s *Server) RegisterRoutes() {
	boxService := s.cfg.BoxService
	boxHandler := handlers.NewBoxHandler(boxService)

	SetupRoutes(s.router, boxHandler)
}

// Run starts the server
func (s *Server) Run() error {
	s.srv = &http.Server{
		Addr:    s.cfg.Port,
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
