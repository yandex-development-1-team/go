package bot

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"
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
	limiters *lru.Cache[int64, *rate.Limiter]
	msgRPS   float64
}

func NewMsgRateLimiter(cacheSize int, msgRPS float64) (*MsgRateLimiter, error) {
	l, err := lru.New[int64, *rate.Limiter](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}
	return &MsgRateLimiter{
		limiters: l,
		msgRPS:   msgRPS,
	}, nil
}

func (rl *MsgRateLimiter) Exec(ctx context.Context, chatID int64, f func() error) error {
	l := rate.NewLimiter(rate.Limit(rl.msgRPS), 1)
	oldL, ok, evicted := rl.limiters.PeekOrAdd(chatID, l)
	if evicted {
		logger.Debug("rate limiter evicted from cache")
	}

	if ok {
		l = oldL
		// check limit exceeded
		if l.Tokens() < 1 {
			metrics.IncBotRateLimit()
			logger.Warn("message limit has been exceeded", zap.Int64("chatID", chatID))
		}
	}
	
	if err := l.Wait(ctx); err != nil {
		return err
	}
	return f()
}
