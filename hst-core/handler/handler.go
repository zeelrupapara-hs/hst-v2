// Package handler is the service itself: app.go builds the world and hands it here.
package handler

import (
	"context"
	"os"
	"runtime"
	"sync"

	"hstcore/config"
	"hstcore/internal/book"
	"hstcore/internal/quote"
	"hstcore/internal/settings"
	"hstcore/internal/shardmap"
	"hstcore/internal/worker"
	"hstcore/model"
	"hstcore/pkg/db"
	"hstcore/pkg/logger"
	"hstcore/pkg/nats"
	"hstcore/pkg/redis"

	natscore "github.com/nats-io/nats.go"
)

// NumCPU sizes the worker pools.
var NumCPU = runtime.NumCPU()

// Handler is the service entry point; every dependency is injected from app.go.
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

	// Quotes is the last price for every symbol. Every pod holds every price.
	Quotes *quote.Book
	// Settings is groups, symbols and the group's overrides of them, already folded together.
	Settings *settings.Store
	// Accounts is the slice of accounts this pod is responsible for.
	Accounts *book.Book

	// Shards says which accounts belong here. Rebuilt whenever the set of live pods changes.
	Shards *shardmap.Map
	// name is how this pod is known on the ring
	name string

	// tempId hands out the keys a position is held under until its row exists
	tempId int64

	// commissions by group id, loaded with the rest of the configuration
	commissions map[int64][]Commission

	// rules is the routing list in evaluation order, replaced wholesale on reload
	rules []model.RoutingRule

	// subs are unsubscribed on Stop
	subs []*natscore.Subscription

	// dealing is the request queue this pod is working, by request id
	dealing map[string]*Pending
	// dealingMu guards the queue on its own, off the rule list's lock
	dealingMu sync.Mutex

	// mu guards the rule list, which is swapped wholesale on reload
	mu sync.RWMutex

	// stop cancels everything Start launched
	stop context.CancelFunc
	wg   sync.WaitGroup
	once sync.Once
}

// New wires the handler; nothing is started here.
func New(cfg *config.Config, log *logger.Logger, database *db.PostgresDB, nc *nats.Nats, rds *redis.Redis) *Handler {
	return &Handler{
		Cfg:      cfg,
		Log:      log,
		DB:       database,
		Nats:     nc,
		Redis:    rds,
		Workers:  worker.New(NumCPU, log),
		Quotes:   quote.New(),
		Settings: settings.New(),
		Accounts: book.New(),
		dealing:  make(map[string]*Pending, 64),
		name:     podName(),
	}
}

// Start loads what the service needs, subscribes, and launches the background goroutines.
func (h *Handler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	h.stop = cancel

	if err := h.load(ctx); err != nil {
		cancel()
		return err
	}

	h.Workers.Start(ctx)

	h.Go(func() { h.runDaily(ctx) })
	h.Go(func() { h.runMembership(ctx) })

	// subscribe last: no message should arrive before the state it reads
	if err := h.subscribe(); err != nil {
		cancel()
		return err
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "handler started", "workers", NumCPU)
	return nil
}

// Stop unsubscribes, cancels the background work and waits. Safe to call more than once.
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

// Subscribe registers a handler and records it for shutdown.
func (h *Handler) Subscribe(subject string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.Subscribe(subject, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

	h.Log.Log(logger.TypeNet, logger.CodeOK, "watching subject", "subject", subject)
	return nil
}

// subscribeQuiet is Subscribe without the log line, for the shard subjects.
func (h *Handler) subscribeQuiet(subject string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.Subscribe(subject, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

	return nil
}

// QueueSubscribe is Subscribe with a queue group.
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

// Go runs f in a goroutine the handler waits for on Stop.
func (h *Handler) Go(f func()) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		f()
	}()
}

// podName is how this instance is known on the shard ring.
func podName() string {
	if n := os.Getenv("POD_NAME"); n != "" {
		return n
	}

	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}

	return "hst-core"
}
