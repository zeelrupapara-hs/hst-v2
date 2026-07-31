package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"hstquote/internal/configclient"
	"hstquote/internal/fixconfig"
	"hstquote/internal/provider"
	"hstquote/internal/provider/fixquotes"
	"hstquote/internal/provider/simulator"
	"hstquote/internal/status"
	"hstquote/internal/tickcache"
	"hstquote/internal/translate"
	"hstquote/model"
	"hstquote/pkg/influxdb"
	"hstquote/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

type streamConnector interface {
	Type() provider.ConnectorType
	Run(ctx context.Context) error
	Close() error
}

type feedRunner struct {
	cancel context.CancelFunc
	conn   streamConnector
}

// Feeds manages quote ingestion loops for configured datafeeds.
type Feeds struct {
	h      *Handler
	config *configclient.Client
	cache  *tickcache.Store
	status *status.Publisher
	influx *influxdb.Client

	mu      sync.Mutex
	runners map[int64]feedRunner
}

func newFeeds(h *Handler) *Feeds {
	return &Feeds{
		h:       h,
		config:  configclient.New(h.Cfg.Quote.ServerURL, h.Cfg.Quote.ServiceToken),
		cache:   tickcache.New(h.Redis),
		status:  status.New(h.Nats, h.Cfg.Nats.Name),
		influx:  h.Influx,
		runners: make(map[int64]feedRunner),
	}
}

func (f *Feeds) ownsFeed(datafeedID int64) bool {
	count := f.h.Cfg.Quote.InstanceCount
	if count <= 1 {
		return true
	}
	idx := f.h.Cfg.Quote.InstanceIndex
	return datafeedID%int64(count) == int64(idx)
}

func (f *Feeds) reload(ctx context.Context) error {
	feeds, err := f.config.ListQuoteFeeds(ctx)
	if err != nil {
		return err
	}

	want := make(map[int64]model.QuoteFeed, len(feeds))
	for _, feed := range feeds {
		if !feed.IsQuoteEnabled() {
			continue
		}
		if !f.ownsFeed(feed.Datafeed.DatafeedID) {
			continue
		}
		if _, ok := provider.ModuleType(feed.Datafeed.Module); !ok {
			f.h.Log.Log(logger.TypeSys, logger.CodeWarn, "unsupported quote module, skipping",
				"datafeed_id", feed.Datafeed.DatafeedID, "module", feed.Datafeed.Module)
			continue
		}
		if len(feed.Translates) == 0 {
			f.h.Log.Log(logger.TypeSys, logger.CodeWarn, "quote feed has no translates, skipping",
				"datafeed_id", feed.Datafeed.DatafeedID)
			continue
		}
		want[feed.Datafeed.DatafeedID] = feed
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for id, runner := range f.runners {
		if _, keep := want[id]; !keep {
			runner.cancel()
			if runner.conn != nil {
				_ = runner.conn.Close()
			}
			f.status.Disconnected(id)
			delete(f.runners, id)
		}
	}

	for id, feed := range want {
		if _, running := f.runners[id]; running {
			continue
		}
		runCtx, cancel := context.WithCancel(ctx)
		feedCopy := feed
		conn := f.newConnector(feedCopy)
		if conn == nil {
			cancel()
			continue
		}
		f.runners[id] = feedRunner{cancel: cancel, conn: conn}
		f.status.Connected(feedCopy.Datafeed.DatafeedID)
		f.h.Go(func() {
			if err := conn.Run(runCtx); err != nil && runCtx.Err() == nil {
				f.h.Log.Log(logger.TypeNet, logger.CodeErr, "quote feed stopped",
					"datafeed_id", feedCopy.Datafeed.DatafeedID, "error", err.Error())
			}
		})
	}

	f.h.Log.Log(logger.TypeSys, logger.CodeOK, "quote feeds synced", "active", len(f.runners))
	return nil
}

func (f *Feeds) reloadOne(ctx context.Context, datafeedID int64) error {
	if !f.ownsFeed(datafeedID) {
		f.mu.Lock()
		if runner, ok := f.runners[datafeedID]; ok {
			runner.cancel()
			if runner.conn != nil {
				_ = runner.conn.Close()
			}
			delete(f.runners, datafeedID)
		}
		f.mu.Unlock()
		f.status.Disconnected(datafeedID)
		return nil
	}

	feed, err := f.config.GetQuoteFeed(ctx, datafeedID)
	if err != nil {
		return err
	}

	f.mu.Lock()
	if runner, ok := f.runners[datafeedID]; ok {
		runner.cancel()
		if runner.conn != nil {
			_ = runner.conn.Close()
		}
		delete(f.runners, datafeedID)
	}
	f.mu.Unlock()
	f.status.Disconnected(datafeedID)

	if feed == nil || !feed.IsQuoteEnabled() || len(feed.Translates) == 0 {
		return nil
	}
	if _, ok := provider.ModuleType(feed.Datafeed.Module); !ok {
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	conn := f.newConnector(*feed)
	if conn == nil {
		cancel()
		return nil
	}

	f.mu.Lock()
	f.runners[datafeedID] = feedRunner{cancel: cancel, conn: conn}
	f.mu.Unlock()
	f.status.Connected(datafeedID)

	feedCopy := *feed
	f.h.Go(func() {
		if err := conn.Run(runCtx); err != nil && runCtx.Err() == nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "quote feed stopped",
				"datafeed_id", feedCopy.Datafeed.DatafeedID, "error", err.Error())
		}
	})
	return nil
}

func (f *Feeds) stopAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for id, runner := range f.runners {
		runner.cancel()
		if runner.conn != nil {
			_ = runner.conn.Close()
		}
		f.status.Disconnected(id)
		delete(f.runners, id)
	}
}

func (f *Feeds) newConnector(feed model.QuoteFeed) streamConnector {
	tickCh := make(chan provider.RawTick, 256)
	typ, _ := provider.ModuleType(feed.Datafeed.Module)

	switch typ {
	case provider.TypeSimulator:
		return &simConnector{
			inner: simulator.NewConnector(feed, tickCh),
			feed:  feed,
			ticks: tickCh,
			f:     f,
		}
	case provider.TypeFIX:
		settings, err := fixconfig.FromFeed(feed)
		if err != nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "fix settings build failed",
				"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
			return nil
		}
		cfgPath, err := fixconfig.ResolveConfigPath(f.h.Cfg.Quote.FixConfigDir, feed, settings)
		if err != nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "fix config build failed",
				"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
			return nil
		}
		return &fixConnector{
			inner: fixquotes.NewConnector(settings, cfgPath, fixconfig.ExternalSymbols(feed), f.h.Log, tickCh),
			feed:  feed,
			ticks: tickCh,
			f:     f,
		}
	default:
		return nil
	}
}

// fixConnector wraps FIX and forwards ticks to the feed handler.
type fixConnector struct {
	inner *fixquotes.Connector
	feed  model.QuoteFeed
	ticks <-chan provider.RawTick
	f     *Feeds
}

func (c *fixConnector) Type() provider.ConnectorType { return c.inner.Type() }

func (c *fixConnector) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.inner.Run(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			_ = c.inner.Close()
			return ctx.Err()
		case err := <-errCh:
			return err
		case raw, ok := <-c.ticks:
			if !ok {
				return nil
			}
			c.f.handleRawTick(ctx, c.feed, raw)
		}
	}
}

func (c *fixConnector) Close() error { return c.inner.Close() }

type simConnector struct {
	inner *simulator.Connector
	feed  model.QuoteFeed
	ticks <-chan provider.RawTick
	f     *Feeds
}

func (c *simConnector) Type() provider.ConnectorType { return c.inner.Type() }

func (c *simConnector) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.inner.Run(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			_ = c.inner.Close()
			return ctx.Err()
		case err := <-errCh:
			return err
		case raw, ok := <-c.ticks:
			if !ok {
				return nil
			}
			c.f.handleRawTick(ctx, c.feed, raw)
		}
	}
}

func (c *simConnector) Close() error { return c.inner.Close() }

func (f *Feeds) handleRawTick(ctx context.Context, feed model.QuoteFeed, raw provider.RawTick) {
	tick, ok := translate.ApplyMarkup(feed, raw)
	if !ok {
		return
	}

	if err := f.cache.Put(ctx, *tick); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "tick cache failed",
			"datafeed_id", tick.DatafeedID, "symbol", tick.Symbol, "error", err.Error())
	}

	f.publishTick(*tick)
	f.status.Tick(tick.DatafeedID, raw.BytesRead)
	if f.influx != nil {
		f.influx.WriteTick(*tick)
	}
}

func (f *Feeds) publishTick(tick model.Tick) {
	payload, err := json.Marshal(tick)
	if err != nil {
		return
	}
	subject := fmt.Sprintf(SubjectTick, tick.Symbol)
	if err := f.h.Nats.NC.Publish(subject, payload); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "tick publish failed",
			"subject", subject, "error", err.Error())
	}
}

func (f *Feeds) onDatafeedEvent(msg *natscore.Msg) {
	var evt DatafeedEvent
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "datafeed event decode failed",
			"error", err.Error())
		return
	}

	if msg.Subject == SubjectDatafeedDeleted || !evt.HasQuoteFlag() {
		f.mu.Lock()
		if runner, ok := f.runners[evt.DatafeedID]; ok {
			runner.cancel()
			if runner.conn != nil {
				_ = runner.conn.Close()
			}
			delete(f.runners, evt.DatafeedID)
		}
		f.mu.Unlock()
		f.status.Disconnected(evt.DatafeedID)
		return
	}

	if err := f.reloadOne(context.Background(), evt.DatafeedID); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "datafeed reload failed",
			"datafeed_id", evt.DatafeedID, "error", err.Error())
	}
}

func (f *Feeds) onConfigSnapshot(msg *natscore.Msg) {
	var snap configclient.Snapshot
	if err := json.Unmarshal(msg.Data, &snap); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "config snapshot decode failed",
			"error", err.Error())
		return
	}
	if snap.Enable != model.DatafeedEnable_enabled || (snap.Mode&model.FeederFlags_quotes) == 0 {
		f.mu.Lock()
		if runner, ok := f.runners[snap.DatafeedID]; ok {
			runner.cancel()
			if runner.conn != nil {
				_ = runner.conn.Close()
			}
			delete(f.runners, snap.DatafeedID)
		}
		f.mu.Unlock()
		f.status.Disconnected(snap.DatafeedID)
		return
	}
	if err := f.reloadOne(context.Background(), snap.DatafeedID); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "config snapshot reload failed",
			"datafeed_id", snap.DatafeedID, "error", err.Error())
	}
}
