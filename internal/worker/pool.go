// Package worker provides a concurrency-safe goroutine worker pool for
// async background job processing (notifications, invoices, analytics).
package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
)

const (
	maxRetries    = 3
	baseBackoffMs = 100
)

// Job is implemented by any value that can execute an async background task.
type Job interface {
	Execute(ctx context.Context)
	Name() string
}

// Pool manages a bounded set of goroutine workers that drain a shared job queue.
type Pool struct {
	jobs        chan Job
	workerCount int
	wg          sync.WaitGroup
	once        sync.Once
}

// NewPool constructs a Pool with the given worker count and buffered queue capacity.
func NewPool(workerCount, queueCapacity int) *Pool {
	return &Pool{
		jobs:        make(chan Job, queueCapacity),
		workerCount: workerCount,
	}
}

// Start launches the worker goroutines. Must be called once before Dispatch.
func (p *Pool) Start(ctx context.Context) {
	for i := range p.workerCount {
		p.wg.Add(1)
		go p.runWorker(ctx, i+1)
	}
	logger.Info(ctx, "worker pool started", "workers", p.workerCount, "queue_capacity", cap(p.jobs))
}

// Dispatch enqueues a job for async processing. Non-blocking; drops the job if
// the queue is full to avoid blocking the HTTP request path.
func (p *Pool) Dispatch(ctx context.Context, job Job) {
	select {
	case p.jobs <- job:
		metrics.WorkerJobsDispatched.WithLabelValues(job.Name()).Inc()
	default:
		logger.Warn(ctx, "worker pool queue full, dropping job", "job", job.Name())
	}
}

// Shutdown drains remaining jobs then waits for all workers to finish.
// Safe to call multiple times; idempotent via sync.Once.
func (p *Pool) Shutdown(ctx context.Context) {
	p.once.Do(func() {
		close(p.jobs)
		logger.Info(ctx, "worker pool shutting down, draining remaining jobs...")
		p.wg.Wait()
		logger.Info(ctx, "worker pool shutdown complete")
	})
}

// runWorker is the long-running goroutine that pulls and executes jobs.
func (p *Pool) runWorker(ctx context.Context, id int) {
	defer p.wg.Done()

	for job := range p.jobs {
		p.executeWithRetry(ctx, id, job)
	}
}

// executeWithRetry runs a job up to maxRetries times with exponential backoff.
// On all retries exhausted, the job is sent to the dead-letter log (DLQ).
func (p *Pool) executeWithRetry(ctx context.Context, workerID int, job Job) {
	var lastErr any

	for attempt := 1; attempt <= maxRetries; attempt++ {
		success := p.executeJob(ctx, workerID, job, attempt)
		if success {
			metrics.WorkerJobsProcessed.WithLabelValues(job.Name()).Inc()
			return
		}

		// Exponential backoff before next retry (skip on last attempt)
		if attempt < maxRetries {
			backoff := time.Duration(baseBackoffMs*(1<<(attempt-1))) * time.Millisecond
			logger.Warn(ctx, "job failed, retrying",
				"worker_id", workerID,
				"job", job.Name(),
				"attempt", attempt,
				"next_retry_in", backoff.String(),
			)
			time.Sleep(backoff)
		}
		lastErr = fmt.Sprintf("attempt %d failed", attempt)
	}

	// All retries exhausted → Dead Letter Queue (DLQ) log
	metrics.WorkerJobsFailed.WithLabelValues(job.Name()).Inc()
	logger.Error(ctx, "job sent to DLQ after all retries exhausted",
		"worker_id", workerID,
		"job", job.Name(),
		"attempts", maxRetries,
		"last_error", fmt.Sprintf("%v", lastErr),
		"dlq", true,
	)
}

// executeJob runs a single job with panic recovery.
// Returns true on success, false if a panic occurred.
func (p *Pool) executeJob(ctx context.Context, workerID int, job Job, attempt int) (success bool) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			logger.Error(ctx, "panic recovered in worker goroutine",
				"worker_id", workerID,
				"job", job.Name(),
				"attempt", attempt,
				"error", fmt.Sprintf("%v", r),
				"stack", stack,
			)
			success = false
		}
	}()

	logger.Info(ctx, "worker executing job", "worker_id", workerID, "job", job.Name(), "attempt", attempt)
	job.Execute(ctx)
	logger.Info(ctx, "worker completed job", "worker_id", workerID, "job", job.Name())
	return true
}
