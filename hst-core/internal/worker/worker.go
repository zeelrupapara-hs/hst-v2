// Package worker is a fixed pool of goroutines over one job channel.
package worker

import (
	"context"
	"sync"

	"hstcore/pkg/logger"
)

// Job is one unit of work.
type Job func(ctx context.Context)

// Pool runs jobs on a fixed number of goroutines.
type Pool struct {
	size  int
	jobs  chan Job
	log   *logger.Logger
	wg    sync.WaitGroup
	once  sync.Once
	ready bool
}

// New builds a pool of size goroutines.
func New(size int, log *logger.Logger) *Pool {
	if size < 1 {
		size = 1
	}
	return &Pool{
		size: size,
		jobs: make(chan Job, size*64),
		log:  log,
	}
}

// Start launches the goroutines.
func (p *Pool) Start(ctx context.Context) {
	p.ready = true

	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.run(ctx)
	}

	p.log.Log(logger.TypeSys, logger.CodeOK, "worker pool started",
		"workers", p.size, "queue", cap(p.jobs))
}

// Submit queues a job.
func (p *Pool) Submit(job Job) bool {
	if !p.ready || job == nil {
		return false
	}

	select {
	case p.jobs <- job:
		return true
	default:
		p.log.Log(logger.TypeSys, logger.CodeWarn, "worker queue full, job dropped",
			"queue", cap(p.jobs))
		return false
	}
}

// Stop closes the queue and waits for the in-flight jobs to finish.
func (p *Pool) Stop() {
	p.once.Do(func() {
		if !p.ready {
			return
		}
		close(p.jobs)
		p.wg.Wait()

		p.log.Log(logger.TypeSys, logger.CodeOK, "worker pool stopped")
	})
}

// Pending is the queue depth, worth exporting as a metric.
func (p *Pool) Pending() int { return len(p.jobs) }

func (p *Pool) run(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				// the queue is closed and drained
				return
			}
			p.exec(ctx, job)

		case <-ctx.Done():
			return
		}
	}
}

// exec keeps one panicking job from taking the whole pool down with it.
func (p *Pool) exec(ctx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			p.log.Log(logger.TypeSys, logger.CodeAtt, "worker recovered from panic",
				"panic", r)
		}
	}()

	job(ctx)
}
