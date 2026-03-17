package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/api/handlers"
	"net/http"

	pgrepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type Server struct {
	router *gin.Engine
}

type Config struct {
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
}

func NewServer(db *sqlx.DB, cfg Config) *Server {
	router := gin.Default()

	rtRepo := pgrepo.NewRefreshTokenRepo(db)
	userRepo := pgrepo.NewUserRepo(db)

	authSvc := svcapi.NewAuthService(
		db,
		rtRepo,
		userRepo,
		cfg.JWTSecret,
		cfg.AccessTokenTTLMinutes,
		cfg.RefreshTokenTTLDays,
	)
	authHandler := handlers.NewAuthHandler(authSvc)

	router.POST("/api/v1/auth/refresh", authHandler.Refresh)
	router.POST("/api/v1/auth/logout", authHandler.Logout)

	return &Server{router: router}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
