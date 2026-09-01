package resilience

import (
	"context"
	"time"
)

type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, burst),
		interval: time.Second / time.Duration(rate),
	}

	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	go rl.refill()

	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
		}
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) Close() {
	close(rl.tokens)
}
