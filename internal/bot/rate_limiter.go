package bot

import (
	"context"
	"sync"

	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	msgRPS = 30.0 // per-chat message limit
	apiRPS = 10.0 // general message limit
)

type ApiRateLimiter struct {
	limiter *rate.Limiter
}

func NewApiRateLimiter() *ApiRateLimiter {
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
	lock     sync.RWMutex
}

func NewMsgRateLimiter() *MsgRateLimiter {
	return &MsgRateLimiter{
		limiters: map[int64]*rate.Limiter{},
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
			l = rate.NewLimiter(msgRPS, 1)
			rl.limiters[chatID] = l
		}
		rl.lock.Unlock()
	}

	// check limit exceeded
	if l.Tokens() < 1 {
		// место для метрики
		logger.Warn("message limit has been exceeded", zap.Int64("chatID", chatID))
	}

	if err := l.Wait(ctx); err != nil {
		return err
	}
	return f()
}
