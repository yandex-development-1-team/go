package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yandex-development-1-team/go/internal/config"
)

const (
	PREFIX = "bot_"
)

var (
	registry *prometheus.Registry

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
	ActiveUsers prometheus.GaugeVec

	// Глобальные лейблы приложения
	appLabels   prometheus.Labels
	labelsNames []string

	initOnce sync.Once
)

func Initialize(cfg config.Config) {
	initOnce.Do(func() {
		initializeMetrics(&cfg)
	})
}

func initializeMetrics(cfg *config.Config) {
	registry = prometheus.NewRegistry()

	// Настраиваем глобальные лейблы при старте
	appLabels = prometheus.Labels{
		"environment": cfg.Environment,
		"instance":    cfg.HostName,
	}
	labelsNames = []string{"environment", "instance"}

	// Инициализация Counter метрик
	MessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_received_total",
			Help: "Total messages received",
		},
		labelsNames,
	)

	MessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_processed_total",
			Help: "Total messages processed",
		},
		labelsNames,
	)

	MessagesErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_errors_total",
			Help: "Total errors during message processing",
		},
		labelsNames,
	)

	DatabaseQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "database_queries_total",
			Help: "Total database queries",
		},
		labelsNames,
	)

	APIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "api_requests_total",
			Help: "Total requests to Telegram API",
		},
		labelsNames,
	)

	BookingsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "bookings_total",
			Help: "Total bookings",
		},
		labelsNames,
	)

	// Инициализация Histogram метрик
	MessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "message_processing_duration_seconds",
			Help:    "Time spent processing messages",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		labelsNames,
	)

	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "database_query_duration_seconds",
			Help:    "Time spent on database queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		labelsNames,
	)

	// Инициализация Gauge метрики
	ActiveUsers = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: PREFIX + "active_users",
			Help: "Number of active users",
		}, labelsNames)

	// Регистрация всех метрик
	registry.MustRegister(MessagesReceived)
	registry.MustRegister(MessagesProcessed)
	registry.MustRegister(MessagesErrors)
	registry.MustRegister(DatabaseQueries)
	registry.MustRegister(APIRequests)
	registry.MustRegister(BookingsTotal)
	registry.MustRegister(MessageProcessingDuration)
	registry.MustRegister(DatabaseQueryDuration)
	registry.MustRegister(ActiveUsers)

	// Стандартные метрики
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// "Up" метрика
	Up := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: PREFIX + "up",
		Help: "Service health status",
	})
	Up.Set(1)
	registry.MustRegister(Up)

}

func NewHandler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
