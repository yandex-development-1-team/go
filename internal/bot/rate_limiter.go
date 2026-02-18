package bot

import (
	"context"
	"sync"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type ApiRateLimiter struct {
	limiter *rate.Limiter
}

func NewApiRateLimiter(apiRPS float64) *ApiRateLimiter {
	return &ApiRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(apiRPS), 1),
	}
}

func (rl *ApiRateLimiter) Exec(ctx context.Context, f func()) error {
	if err := rl.limiter.Wait(ctx); err != nil {
		return err
	}
	f()
	return nil
}

type MsgRateLimiter struct {
	limiters map[int64]*rate.Limiter
	msgRPS   float64
	lock     sync.RWMutex
}

func NewMsgRateLimiter(msgRPS float64) *MsgRateLimiter {
	return &MsgRateLimiter{
		limiters: map[int64]*rate.Limiter{},
		msgRPS:   msgRPS,
	}
}

func (rl *MsgRateLimiter) Exec(ctx context.Context, chatID int64, f func() error) error {
	rl.lock.RLock()
	l := rl.limiters[chatID]
	rl.lock.RUnlock()
	if l == nil {
		rl.lock.Lock()
		l = rl.limiters[chatID]
		// double check, to avoid race
		if l == nil {
			l = rate.NewLimiter(rate.Limit(rl.msgRPS), 1)
			rl.limiters[chatID] = l
		}
		rl.lock.Unlock()
	}

	// check limit exceeded
	if l.Tokens() < 1 {
		metrics.IncBotRateLimit()
		logger.Warn("message limit has been exceeded", zap.Int64("chatID", chatID))
	}

	if err := l.Wait(ctx); err != nil {
		return err
	}
	return f()
}
