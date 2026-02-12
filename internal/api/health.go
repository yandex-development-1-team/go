package api

type healthHandler struct {
	db     *sqlx.DB
	logger *zap.Logger

	// кэш для Telegram
	tgMu     sync.RWMutex
	tgStatus healthCheckResult
	tgLast   time.Time
}

type healthCheckResult struct {
	Status string    `json:"status"`
	Error  error     `json:"-"`
	Time   time.Time `json:"-"`
}
