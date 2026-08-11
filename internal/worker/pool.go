// Package worker provides a concurrency-safe goroutine worker pool for
// async background job processing (notifications, invoices, analytics).
package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
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
		p.executeJob(ctx, id, job)
	}
}

// executeJob runs a single job with panic recovery so one bad job cannot crash a worker.
func (p *Pool) executeJob(ctx context.Context, workerID int, job Job) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			logger.Error(ctx, "panic recovered in worker goroutine",
				"worker_id", workerID,
				"job", job.Name(),
				"error", fmt.Sprintf("%v", r),
				"stack", stack,
			)
		}
	}()

	logger.Info(ctx, "worker executing job", "worker_id", workerID, "job", job.Name())
	job.Execute(ctx)
	metrics.WorkerJobsProcessed.WithLabelValues(job.Name()).Inc()
	logger.Info(ctx, "worker completed job", "worker_id", workerID, "job", job.Name())
}
