package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errRateLimiterClosed = errors.New("rate limiter closed")

type RateLimiter struct {
	tokens    chan struct{}
	interval  time.Duration
	stop      chan struct{}
	stopped   chan struct{}
	closeOnce sync.Once
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, burst),
		interval: time.Second / time.Duration(rate),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
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
	defer close(rl.stopped)

	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			select {
			case <-rl.stop:
				return
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.stop:
		return errRateLimiterClosed
	default:
	}

	select {
	case <-rl.tokens:
		return nil
	case <-rl.stop:
		return errRateLimiterClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) Close() {
	rl.closeOnce.Do(func() {
		close(rl.stop)
		<-rl.stopped
	})
}
