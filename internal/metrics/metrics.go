package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yandex-development-1-team/go/internal/config"
)

const (
	PREFIX = "bot_"
)

var (
	Registry *prometheus.Registry

	// Counter метрики
	MessagesReceived  *prometheus.CounterVec
	MessagesProcessed *prometheus.CounterVec
	MessagesErrors    *prometheus.CounterVec
	DatabaseQueries   *prometheus.CounterVec
	APIRequests       *prometheus.CounterVec
	BookingsTotal     *prometheus.CounterVec

	// Histogram метрики
	MessageProcessingDuration *prometheus.HistogramVec
	DatabaseQueryDuration     *prometheus.HistogramVec

	// Gauge метрика
	ActiveUsers prometheus.Gauge

	// Глобальные лейблы приложения
	appLabels  prometheus.Labels
	labelNames []string
)

func Initialize(cfg config.Config) {
	Registry = prometheus.NewRegistry()

	// Настраиваем глобальные лейблы при старте
	appLabels = prometheus.Labels{
		"environment": cfg.Environment,
		"instance":    cfg.HostName,
	}
	labelNames = []string{"environment", "instance"}

	// Инициализация Counter метрик
	MessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_received_total",
			Help: "Total messages received",
		},
		[]string{}, // без лейблов
	)

	MessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_processed_total",
			Help: "Total messages processed",
		},
		[]string{},
	)

	MessagesErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_errors_total",
			Help: "Total errors during message processing",
		},
		[]string{},
	)

	DatabaseQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "database_queries_total",
			Help: "Total database queries",
		},
		[]string{},
	)

	APIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "api_requests_total",
			Help: "Total requests to Telegram API",
		},
		[]string{},
	)

	BookingsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "bookings_total",
			Help: "Total bookings",
		},
		[]string{},
	)

	// Инициализация Histogram метрик
	MessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "message_processing_duration_seconds",
			Help:    "Time spent processing messages",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{},
	)

	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "database_query_duration_seconds",
			Help:    "Time spent on database queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{},
	)

	// Инициализация Gauge метрики
	ActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: PREFIX + "active_users",
			Help: "Number of active users",
		},
	)

	// Регистрация всех метрик
	Registry.MustRegister(MessagesReceived)
	Registry.MustRegister(MessagesProcessed)
	Registry.MustRegister(MessagesErrors)
	Registry.MustRegister(DatabaseQueries)
	Registry.MustRegister(APIRequests)
	Registry.MustRegister(BookingsTotal)
	Registry.MustRegister(MessageProcessingDuration)
	Registry.MustRegister(DatabaseQueryDuration)
	Registry.MustRegister(ActiveUsers)
}

// StartMetricsServer запускает HTTP сервер для /metrics
func StartMetricsServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))

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
