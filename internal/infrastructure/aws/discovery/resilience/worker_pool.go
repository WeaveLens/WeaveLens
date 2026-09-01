package resilience

import (
	"context"
	"sync"
)

type Task func(ctx context.Context) error

type WorkerPool struct {
	workers int
	tasks   chan Task
	wg      sync.WaitGroup
	mu      sync.Mutex
	errs    []error
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		tasks:   make(chan Task, workers),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			if err := task(ctx); err != nil {
				wp.mu.Lock()
				wp.errs = append(wp.errs, err)
				wp.mu.Unlock()
			}
		}
	}
}

func (wp *WorkerPool) Submit(task Task) {
	wp.tasks <- task
}

func (wp *WorkerPool) Stop() []error {
	close(wp.tasks)
	wp.wg.Wait()

	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.errs
}
