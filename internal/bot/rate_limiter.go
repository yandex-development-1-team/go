package bot

import (
	"context"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps float64) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
	}
}

func (rl *RateLimiter) Exec(ctx context.Context, f func() error) error {
	if err := rl.limiter.Wait(ctx); err != nil {
		return err
	}
	return f()
}
