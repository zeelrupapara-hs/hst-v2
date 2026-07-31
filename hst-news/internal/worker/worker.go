// Package worker is a fixed pool of goroutines over one job channel.
//
// Spawning a goroutine per message is fine until the day traffic spikes and
// the process is holding a hundred thousand of them. A fixed pool bounds that:
// the queue grows instead of the goroutine count, and a full queue is a signal
// the service is behind rather than a slow death by scheduler.
package worker

import (
	"context"
	"sync"

	"hstnews/pkg/logger"
)

// Job is one unit of work. It gets the pool's context so a long job can notice
// shutdown instead of holding it up.
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

// New builds a pool of size goroutines. The queue is deliberately deeper than
// the pool, so a short burst is absorbed rather than rejected.
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

// Start launches the goroutines. They exit when ctx is cancelled or Stop is
// called, whichever happens first.
func (p *Pool) Start(ctx context.Context) {
	p.ready = true

	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.run(ctx)
	}

	p.log.Log(logger.TypeSys, logger.CodeOK, "worker pool started",
		"workers", p.size, "queue", cap(p.jobs))
}

// Submit queues a job. It reports false when the queue is full rather than
// blocking the caller, which is usually a nats callback that must not stall.
// A false return is a real signal: log it, count it, shed load.
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

// Stop closes the queue and waits for the in-flight jobs to finish. Queued
// jobs still run: they were accepted, so dropping them would be a lie.
// Safe to call more than once.
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

// Pending is the queue depth, worth exporting as a metric: it is the first
// number that moves when the service starts falling behind.
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
