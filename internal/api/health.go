package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

type healthHandler struct {
	db    *sqlx.DB
	tgAPI string

	// Telegram cache
	tgMu     sync.RWMutex
	tgStatus healthCheckResult
	tgLast   time.Time
}

type healthCheckResult struct {
	Status string    `json:"status"`
	Error  error     `json:"-"`
	Time   time.Time `json:"-"`
}

type healthCheckResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	DB        string    `json:"db"`
	Telegram  string    `json:"telegram"`
}

const (
	checkTimeout = 2 * time.Second
)

func NewHealthHandler(db *sqlx.DB, apiUrl string) http.HandlerFunc {
	h := &healthHandler{db: db, tgAPI: apiUrl}
	return h.Health
}

func (h *healthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbRes := h.checkDB(ctx)
	tgRes := h.checkTelegram(ctx)

	logger.Info("health check",
		zap.String("handler", "health"),
		zap.String("db", dbRes.Status),
		zap.String("telegram", tgRes.Status),
		zap.Error(dbRes.Error),
		zap.NamedError("telegram_error", tgRes.Error),
	)

	status := "ok"
	code := http.StatusOK
	if dbRes.Status != "ok" || tgRes.Status != "ok" {
		status = "fail"
		code = http.StatusServiceUnavailable
	}

	resp := healthCheckResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		DB:        dbRes.Status,
		Telegram:  tgRes.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("failed to encode health response", zap.String("handler", "health"), zap.Error(err))
	}

}

func (h *healthHandler) checkDB(ctx context.Context) healthCheckResult {

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	if h.db == nil {
		return healthCheckResult{Status: "fail", Error: errors.New("db is not configured"), Time: time.Now()}
	}

	if err := h.db.PingContext(ctx); err != nil {
		return healthCheckResult{Status: "fail", Error: err, Time: time.Now()}
	}
	return healthCheckResult{Status: "ok", Time: time.Now()}
}

func (h *healthHandler) checkTelegram(ctx context.Context) healthCheckResult {
	h.tgMu.RLock()
	if time.Since(h.tgLast) < time.Minute {
		cached := h.tgStatus
		h.tgMu.RUnlock()
		return cached
	}
	h.tgMu.RUnlock()

	h.tgMu.Lock()
	defer h.tgMu.Unlock()

	if time.Since(h.tgLast) < time.Minute {
		return h.tgStatus
	}

	res := h.doTelegramCheck(ctx)
	h.tgStatus = res
	h.tgLast = time.Now()
	return res
}

func (h *healthHandler) doTelegramCheck(ctx context.Context) healthCheckResult {
	const timeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "HEAD", h.tgAPI, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return healthCheckResult{Status: "fail", Error: err, Time: time.Now()}
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return healthCheckResult{Status: "fail", Error: fmt.Errorf("status %d", resp.StatusCode), Time: time.Now()}
	}
	return healthCheckResult{Status: "ok", Time: time.Now()}
}
