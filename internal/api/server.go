package api

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/api/handlers"
	pgrepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type Server struct {
	mux *http.ServeMux
}

type Config struct {
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
}

func NewServer(db *sqlx.DB, cfg Config) *Server {
	mux := http.NewServeMux()

	rtRepo := pgrepo.NewRefreshTokenRepo(db)
	userRepo := pgrepo.NewUserRepo(db)
	txRepo := pgrepo.NewTxRepo(db)

	authSvc := svcapi.NewAuthService(
		db,
		rtRepo,
		userRepo,
		txRepo,
		cfg.JWTSecret,
		cfg.AccessTokenTTLMinutes,
		cfg.RefreshTokenTTLDays,
	)
	authHandler := handlers.NewAuthHandler(authSvc)

	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	return &Server{mux: mux}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
