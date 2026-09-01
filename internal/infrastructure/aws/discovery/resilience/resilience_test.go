package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), DefaultRetryConfig(), nil, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		JitterFactor: 0,
	}

	calls := 0
	err := Retry(context.Background(), cfg, func(err error) bool {
		return true
	}, func() error {
		calls++
		if calls < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		JitterFactor: 0,
	}

	calls := 0
	err := Retry(context.Background(), cfg, func(err error) bool {
		return false
	}, func() error {
		calls++
		return errors.New("non-retryable")
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call for non-retryable error, got %d", calls)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		JitterFactor: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	calls := 0
	err := Retry(ctx, cfg, func(err error) bool {
		return true
	}, func() error {
		calls++
		return errors.New("transient")
	})

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	if calls > 2 {
		t.Errorf("expected early termination, got %d calls", calls)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		JitterFactor: 0,
	}

	var delays []time.Duration
	lastTime := time.Now()

	err := Retry(context.Background(), cfg, func(err error) bool {
		return true
	}, func() error {
		now := time.Now()
		delays = append(delays, now.Sub(lastTime))
		lastTime = now
		return errors.New("transient")
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if len(delays) != 4 {
		t.Fatalf("expected 4 delays, got %d", len(delays))
	}

	for i := 1; i < len(delays)-1; i++ {
		if delays[i] < delays[i-1] {
			t.Errorf("delay %d (%v) should be >= delay %d (%v)", i, delays[i], i-1, delays[i-1])
		}
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(100, 10)
	defer rl.Close()

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait() unexpected error: %v", err)
		}
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	defer rl.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestWorkerPool_ExecutesTasks(t *testing.T) {
	pool := NewWorkerPool(4)
	ctx := context.Background()
	pool.Start(ctx)

	var counter atomic.Int32
	for i := 0; i < 10; i++ {
		pool.Submit(func(ctx context.Context) error {
			counter.Add(1)
			return nil
		})
	}

	errs := pool.Stop()
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
	if counter.Load() != 10 {
		t.Errorf("expected 10 tasks executed, got %d", counter.Load())
	}
}

func TestWorkerPool_CollectsErrors(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	for i := 0; i < 5; i++ {
		pool.Submit(func(ctx context.Context) error {
			return errors.New("task error")
		})
	}

	errs := pool.Stop()
	if len(errs) != 5 {
		t.Errorf("expected 5 errors, got %d", len(errs))
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	cancel()

	errs := pool.Stop()
	_ = errs
}
