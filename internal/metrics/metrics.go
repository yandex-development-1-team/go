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

	// Counter metrics
	messagesReceived  *prometheus.CounterVec
	messagesProcessed *prometheus.CounterVec
	messagesErrors    *prometheus.CounterVec
	databaseQueries   *prometheus.CounterVec
	apiRequests       *prometheus.CounterVec
	bookingsTotal     *prometheus.CounterVec

	// Histogram metrics
	messageProcessingDuration *prometheus.HistogramVec
	databaseQueryDuration     *prometheus.HistogramVec

	// Gauge metrics
	activeUsers prometheus.GaugeVec

	// Global app lables
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

	// init Global lables
	appLabels = prometheus.Labels{
		"environment": cfg.Environment,
		"instance":    cfg.HostName,
	}
	labelsNames = []string{"environment", "instance"}

	// init Counter metrics
	messagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_received_total",
			Help: "Total messages received",
		},
		labelsNames,
	)

	messagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_processed_total",
			Help: "Total messages processed",
		},
		labelsNames,
	)

	messagesErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "messages_errors_total",
			Help: "Total errors during message processing",
		},
		labelsNames,
	)

	databaseQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "database_queries_total",
			Help: "Total database queries",
		},
		labelsNames,
	)

	apiRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "api_requests_total",
			Help: "Total requests to Telegram API",
		},
		labelsNames,
	)

	bookingsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: PREFIX + "bookings_total",
			Help: "Total bookings",
		},
		labelsNames,
	)

	// init Histogram metrics
	messageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "message_processing_duration_seconds",
			Help:    "Time spent processing messages",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		labelsNames,
	)

	databaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    PREFIX + "database_query_duration_seconds",
			Help:    "Time spent on database queries",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		labelsNames,
	)

	// init Gauge metrics
	activeUsers = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: PREFIX + "active_users",
			Help: "Number of active users",
		}, labelsNames)

	// metrics register
	registry.MustRegister(messagesReceived)
	registry.MustRegister(messagesProcessed)
	registry.MustRegister(messagesErrors)
	registry.MustRegister(databaseQueries)
	registry.MustRegister(apiRequests)
	registry.MustRegister(bookingsTotal)
	registry.MustRegister(messageProcessingDuration)
	registry.MustRegister(databaseQueryDuration)
	registry.MustRegister(activeUsers)

	// standart metrics
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// "Up" metric
	up := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: PREFIX + "up",
		Help: "Service health status",
	})
	up.Set(1)
	registry.MustRegister(up)

}

func NewHandler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func IncMessagesReceived() {
	messagesReceived.With(appLabels).Inc()
}

func IncMessagesProcessed() {
	messagesProcessed.With(appLabels).Inc()
}

func IncMessagesErrors() {
	messagesErrors.With(appLabels).Inc()
}

func IncDatabaseQueries() {
	databaseQueries.With(appLabels).Inc()
}

func IncAPIRequests() {
	apiRequests.With(appLabels).Inc()
}

func IncBookingsTotal() {
	bookingsTotal.With(appLabels).Inc()
}

func ObserveMessageProcessingDuration(seconds float64) {
	messageProcessingDuration.With(appLabels).Observe(seconds)
}

func ObserveDatabaseQueryDuration(seconds float64) {
	databaseQueryDuration.With(appLabels).Observe(seconds)
}

func SetActiveUsers(count int) {
	activeUsers.With(appLabels).Set(float64(count))
}
