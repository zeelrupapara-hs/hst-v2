package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"hstcore/config"
	"hstcore/pkg/logger"
)

func newTestPool(t *testing.T, size int) *Pool {
	t.Helper()

	// the logger writes a day file, so keep it inside the test's temp dir
	t.Setenv("LOG_DIR", t.TempDir())

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	log, err := logger.NewLogger(cfg)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	t.Cleanup(func() { _ = log.Close() })

	return New(size, log)
}

func TestPoolRunsEverySubmittedJob(t *testing.T) {
	p := newTestPool(t, 4)
	p.Start(context.Background())

	var done sync.WaitGroup
	var count atomic.Int64

	for i := 0; i < 100; i++ {
		done.Add(1)
		if !p.Submit(func(context.Context) {
			defer done.Done()
			count.Add(1)
		}) {
			done.Done()
			t.Fatal("the queue rejected a job it had room for")
		}
	}

	done.Wait()
	p.Stop()

	if got := count.Load(); got != 100 {
		t.Fatalf("expected 100 jobs, got %d", got)
	}
}

func TestSubmitReportsAFullQueue(t *testing.T) {
	p := newTestPool(t, 1)
	// not started, so nothing drains and the queue fills up
	p.ready = true

	filled := 0
	for i := 0; i < cap(p.jobs)+10; i++ {
		if p.Submit(func(context.Context) {}) {
			filled++
		}
	}

	if filled != cap(p.jobs) {
		t.Fatalf("expected the queue to accept exactly %d, got %d", cap(p.jobs), filled)
	}
}

func TestSubmitBeforeStartIsRejected(t *testing.T) {
	p := newTestPool(t, 1)

	if p.Submit(func(context.Context) {}) {
		t.Fatal("a pool that is not started must not accept work")
	}
}

func TestQueuedJobsStillRunOnStop(t *testing.T) {
	p := newTestPool(t, 1)
	p.Start(context.Background())

	var count atomic.Int64
	for i := 0; i < 20; i++ {
		p.Submit(func(context.Context) { count.Add(1) })
	}

	// an accepted job is a promise, so Stop drains rather than discards
	p.Stop()

	if got := count.Load(); got != 20 {
		t.Fatalf("expected all 20 queued jobs to run, got %d", got)
	}
}

func TestPanickingJobDoesNotKillTheWorker(t *testing.T) {
	p := newTestPool(t, 1)
	p.Start(context.Background())

	p.Submit(func(context.Context) { panic("boom") })

	survived := make(chan struct{})
	p.Submit(func(context.Context) { close(survived) })

	select {
	case <-survived:
	case <-time.After(time.Second):
		t.Fatal("the worker died with the panicking job")
	}

	p.Stop()
}

func TestStopIsIdempotent(t *testing.T) {
	p := newTestPool(t, 2)
	p.Start(context.Background())

	p.Stop()
	p.Stop()
}

func TestCancellingTheContextStopsTheWorkers(t *testing.T) {
	p := newTestPool(t, 2)
	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)

	cancel()

	// Stop must still return once the workers have exited on their own
	stopped := make(chan struct{})
	go func() {
		p.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop hung after the context was cancelled")
	}
}
