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

	// rules is the routing list, in evaluation order. Replaced wholesale on reload, never
	// edited in place, so a request already walking it sees one consistent list.
	rules []model.RoutingRule

	// subs are unsubscribed on Stop, so a shutdown does not leave a consumer
	// attached to a connection that is about to drain
	subs []*natscore.Subscription

	// mu guards the rule list, which is swapped wholesale on reload
	mu sync.RWMutex

	// stop cancels everything Start launched
	stop context.CancelFunc
	wg   sync.WaitGroup
	once sync.Once
}

// New wires the handler. Nothing is started here: a constructor that spawns
// goroutines cannot be used in a test without also shutting it down.
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
		name:     podName(),
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

// subscribe registers every consumer.
//
// Prices are broadcast: every pod hears every symbol, because any pod may hold an account
// trading it. Trade requests are not: a pod hears only the shards it owns, so a request is
// handled once, by the pod that already has the account in memory.
func (h *Handler) subscribe() error {
	if err := h.Subscribe(model.TickSubjectAll, h.onTick); err != nil {
		return err
	}

	for _, shard := range h.Shards.Mine() {
		if err := h.subscribeQuiet(shardmap.SubjectFor(shard), h.onTradeRequest); err != nil {
			return err
		}
	}

	// a manager changing the rules or the groups tells every pod to read them again
	if err := h.Subscribe(SubjectReload, h.onReload); err != nil {
		return err
	}

	h.Log.Log(logger.TypeNet, logger.CodeOK, "engine listening",
		"shards", len(h.Shards.Mine()), "ticks", model.TickSubjectAll)

	return nil
}

// SubjectReload is what the API server publishes when configuration changed underneath us.
const SubjectReload = "system.core.reload"

// onReload reads the configuration again. Cheap enough to do wholesale: it is a handful of
// tables and it happens when somebody clicks save, not on the trading path.
func (h *Handler) onReload(msg *natscore.Msg) {
	ctx := context.Background()

	if err := h.loadSettings(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload settings", "error", err.Error())
		return
	}
	if err := h.loadRules(ctx); err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeErr, "could not reload routing rules", "error", err.Error())
		return
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "configuration reloaded",
		"groups", h.Settings.Groups(), "symbols", h.Settings.Symbols())
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

// subscribeQuiet is Subscribe without the log line. Used for the shard subjects, where there
// is one per shard and the count is the only interesting part.
func (h *Handler) subscribeQuiet(subject string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.Subscribe(subject, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

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

// podName is how this instance is known on the shard ring.
//
// In Kubernetes the pod name is unique and stable for the life of the pod, which is exactly
// what the ring needs. Falling back to the hostname keeps a local run working.
func podName() string {
	if n := os.Getenv("POD_NAME"); n != "" {
		return n
	}

	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}

	return "hst-core"
}
