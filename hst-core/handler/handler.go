// Package handler is the service itself. app.go builds the world — config,
// logger, postgres, nats, redis — and hands it here; everything this service
// actually does hangs off the Handler.
//
// One struct holds every dependency, so a method never reaches for a global
// and a test can build a Handler with fakes. Add a file per concern next to
// this one (accounts.go, orders.go, alerts.go) and keep this file to the
// lifecycle: New, Start, Stop.
package handler

import (
	"context"
	"runtime"
	"sync"

	"hstcore/config"
	"hstcore/internal/worker"
	"hstcore/pkg/db"
	"hstcore/pkg/logger"
	"hstcore/pkg/nats"
	"hstcore/pkg/redis"

	natscore "github.com/nats-io/nats.go"
)

// NumCPU sizes the worker pools. One goroutine per core is the starting point;
// measure before changing it.
var NumCPU = runtime.NumCPU()

// Handler is the service entry point. Every dependency is injected from
// app.go, which keeps the runtime of each microservice isolated.
type Handler struct {
	// Cfg is the validated configuration
	Cfg *config.Config
	// Log writes to stdout and the day file
	Log *logger.Logger
	// DB is the postgres pool
	DB *db.PostgresDB
	// Nats is the message bus
	Nats *nats.Nats
	// Redis is the cache
	Redis *redis.Redis

	// Workers runs the background jobs this service queues
	Workers *worker.Pool

	// subs are unsubscribed on Stop, so a shutdown does not leave a consumer
	// attached to a connection that is about to drain
	subs []*natscore.Subscription

	// stop cancels everything Start launched
	stop context.CancelFunc
	wg   sync.WaitGroup
	once sync.Once
}

// New wires the handler. Nothing is started here: a constructor that spawns
// goroutines cannot be used in a test without also shutting it down.
func New(cfg *config.Config, log *logger.Logger, database *db.PostgresDB, nc *nats.Nats, rds *redis.Redis) *Handler {
	return &Handler{
		Cfg:     cfg,
		Log:     log,
		DB:      database,
		Nats:    nc,
		Redis:   rds,
		Workers: worker.New(NumCPU, log),
	}
}

// Start loads what the service needs, subscribes, and launches the background
// goroutines. It returns as soon as everything is running; the caller blocks
// on a signal, not on this.
func (h *Handler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	h.stop = cancel

	// load state into memory before anything can ask for it
	if err := h.load(ctx); err != nil {
		cancel()
		return err
	}

	h.Workers.Start(ctx)

	// subscribe last: no message should arrive before the state it reads
	if err := h.subscribe(); err != nil {
		cancel()
		return err
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "handler started", "workers", NumCPU)
	return nil
}

// Stop unsubscribes, cancels the background work and waits for it to finish.
// Safe to call more than once.
func (h *Handler) Stop() {
	h.once.Do(func() {
		for _, s := range h.subs {
			if err := s.Unsubscribe(); err != nil {
				h.Log.Log(logger.TypeNet, logger.CodeWarn, "unsubscribe failed",
					"subject", s.Subject, "error", err.Error())
			}
		}

		if h.stop != nil {
			h.stop()
		}
		h.Workers.Stop()
		h.wg.Wait()

		h.Log.Log(logger.TypeSys, logger.CodeOK, "handler stopped")
	})
}

// load reads whatever this service keeps in memory. Called before the first
// subscription, so a handler never sees a half loaded world.
func (h *Handler) load(ctx context.Context) error {
	// TODO: load what this service caches at boot.
	_ = ctx
	return nil
}

// subscribe registers every consumer. Keep the subjects in nats.go.
func (h *Handler) subscribe() error {
	// TODO: register this service's subjects, for example
	//
	//	if err := h.Subscribe(SubjectExample, h.onExample); err != nil {
	//		return err
	//	}
	return nil
}

// Subscribe registers a handler and records it for shutdown. Unlike a Fatal on
// a failed subscribe, this lets the caller decide, and a service that cannot
// hear its own subject should refuse to boot rather than run deaf.
func (h *Handler) Subscribe(subject string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.Subscribe(subject, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

	h.Log.Log(logger.TypeNet, logger.CodeOK, "watching subject", "subject", subject)
	return nil
}

// QueueSubscribe is Subscribe with a queue group, so exactly one instance in
// the group handles each message. This is what horizontal scaling needs.
func (h *Handler) QueueSubscribe(subject, group string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.QueueSubscribe(subject, group, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

	h.Log.Log(logger.TypeNet, logger.CodeOK, "watching subject",
		"subject", subject, "queue", group)
	return nil
}

// Go runs f in a goroutine the handler waits for on Stop. Use it instead of a
// bare go statement so shutdown is not a race.
func (h *Handler) Go(f func()) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		f()
	}()
}
