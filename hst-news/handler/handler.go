// Package handler is the hst-news service: poll configured news datafeeds,
// normalize items, cache in redis, and publish on NATS.
package handler

import (
	"context"
	"runtime"
	"sync"
	"time"

	"hstnews/config"
	"hstnews/internal/worker"
	"hstnews/pkg/logger"
	"hstnews/pkg/nats"
	"hstnews/pkg/redis"

	natscore "github.com/nats-io/nats.go"
)

var NumCPU = runtime.NumCPU()

// Handler is the service entry point.
type Handler struct {
	Cfg     *config.Config
	Log     *logger.Logger
	Nats    *nats.Nats
	Redis   *redis.Redis
	Workers *worker.Pool
	Feeds   *Feeds

	subs []*natscore.Subscription

	stop context.CancelFunc
	wg   sync.WaitGroup
	once sync.Once
}

func New(cfg *config.Config, log *logger.Logger, nc *nats.Nats, rds *redis.Redis) *Handler {
	h := &Handler{
		Cfg:     cfg,
		Log:     log,
		Nats:    nc,
		Redis:   rds,
		Workers: worker.New(NumCPU, log),
	}
	h.Feeds = newFeeds(h)
	return h
}

func (h *Handler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	h.stop = cancel

	if err := h.load(ctx); err != nil {
		cancel()
		return err
	}

	h.Workers.Start(ctx)

	if err := h.subscribe(); err != nil {
		cancel()
		return err
	}

	h.Go(func() { h.runConfigReload(ctx) })

	h.Log.Log(logger.TypeSys, logger.CodeOK, "handler started", "workers", NumCPU)
	return nil
}

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
		h.Feeds.stopAll()
		h.Workers.Stop()
		h.wg.Wait()

		h.Log.Log(logger.TypeSys, logger.CodeOK, "handler stopped")
	})
}

func (h *Handler) load(ctx context.Context) error {
	return h.Feeds.reload(ctx)
}

func (h *Handler) subscribe() error {
	if err := h.QueueSubscribe(SubjectDatafeedCreated, GroupDatafeedConfig, h.Feeds.onDatafeedEvent); err != nil {
		return err
	}
	if err := h.QueueSubscribe(SubjectDatafeedUpdated, GroupDatafeedConfig, h.Feeds.onDatafeedEvent); err != nil {
		return err
	}
	if err := h.QueueSubscribe(SubjectDatafeedDeleted, GroupDatafeedConfig, h.Feeds.onDatafeedEvent); err != nil {
		return err
	}
	if err := h.QueueSubscribe(SubjectDatafeedConfig, GroupDatafeedConfig, h.Feeds.onConfigSnapshot); err != nil {
		return err
	}
	return nil
}

func (h *Handler) runConfigReload(ctx context.Context) {
	interval := h.Cfg.News.ReloadInterval
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := h.Feeds.reload(context.Background()); err != nil {
				h.Log.Log(logger.TypeNet, logger.CodeErr, "periodic feed reload failed",
					"error", err.Error())
			}
		}
	}
}

func (h *Handler) Subscribe(subject string, cb natscore.MsgHandler) error {
	sub, err := h.Nats.NC.Subscribe(subject, cb)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, sub)

	h.Log.Log(logger.TypeNet, logger.CodeOK, "watching subject", "subject", subject)
	return nil
}

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

func (h *Handler) Go(f func()) {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		f()
	}()
}
