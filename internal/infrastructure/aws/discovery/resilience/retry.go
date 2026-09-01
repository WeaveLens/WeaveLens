package resilience

import (
	"context"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	JitterFactor float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		JitterFactor: 0.3,
	}
}

type IsRetryable func(err error) bool

func Retry(ctx context.Context, cfg RetryConfig, isRetryable IsRetryable, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err

			if isRetryable != nil && !isRetryable(err) {
				return err
			}

			if attempt == cfg.MaxAttempts-1 {
				return err
			}

			delay := calculateDelay(cfg, attempt)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
		return nil
	}

	return lastErr
}

func calculateDelay(cfg RetryConfig, attempt int) time.Duration {
	delay := cfg.InitialDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
			break
		}
	}

	if cfg.JitterFactor > 0 {
		jitter := time.Duration(float64(delay) * cfg.JitterFactor * (rand.Float64()*2 - 1))
		delay += jitter
		if delay < 0 {
			delay = cfg.InitialDelay
		}
	}

	return delay
}
