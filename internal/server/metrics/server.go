package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/metrics"
)

type MetricsServer struct {
	server *http.Server
}

func NewServer(cfg config.Config) error {

	metrics.Initialize(cfg)

	return nil
}

// StartMetricsServer запускает HTTP сервер для /metrics
func StartMetricsServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server
}

// StopMetricsServer корректно завершает сервер метрик
func StopMetricsServer(server *http.Server) error {
	return server.Close()
}
