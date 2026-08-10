package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"hstquote/internal/configclient"
	"hstquote/internal/filter"
	"hstquote/internal/fixconfig"
	"hstquote/internal/provider"
	"hstquote/internal/provider/fixquotes"
	"hstquote/internal/session"
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
	done   chan struct{}
}

const runnerStopTimeout = 10 * time.Second

// stopRunnerLocked cancels a feed runner and waits for its goroutine to exit so
// QuickFIX can release the SessionID before a replacement initiator starts.
func (f *Feeds) stopRunnerLocked(id int64, runner feedRunner) {
	runner.cancel()
	if runner.conn != nil {
		_ = runner.conn.Close()
	}
	if runner.done != nil {
		select {
		case <-runner.done:
		case <-time.After(runnerStopTimeout):
			f.h.Log.Log(logger.TypeSys, logger.CodeWarn, "quote feed runner stop timed out",
				"datafeed_id", id)
		}
	}
}

// Feeds manages quote ingestion loops for configured datafeeds.
type Feeds struct {
	h      *Handler
	config *configclient.Client
	cache  *tickcache.Store
	status *status.Publisher
	influx *influxdb.Client

	mu        sync.Mutex
	runners   map[int64]feedRunner
	liveFeeds map[int64]*atomic.Pointer[model.QuoteFeed]
}

func newFeeds(h *Handler) *Feeds {
	return &Feeds{
		h:         h,
		config:    configclient.New(h.Cfg.Quote.ServerURL, h.Cfg.Quote.ServiceToken),
		cache:     tickcache.New(h.Redis),
		status:    status.New(h.Nats, h.Cfg.Nats.Name),
		influx:    h.Influx,
		runners:   make(map[int64]feedRunner),
		liveFeeds: make(map[int64]*atomic.Pointer[model.QuoteFeed]),
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

// bindLiveFeed swaps in a whole new config; runners read it without a lock, so it is never mutated in place.
func (f *Feeds) bindLiveFeed(id int64, feed model.QuoteFeed) *atomic.Pointer[model.QuoteFeed] {
	ptr, ok := f.liveFeeds[id]
	if !ok {
		ptr = &atomic.Pointer[model.QuoteFeed]{}
		f.liveFeeds[id] = ptr
	}
	ptr.Store(&feed)
	return ptr
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
			f.stopRunnerLocked(id, runner)
			f.status.Disconnected(id)
			f.status.Journal(id, status.JournalInfo, "disconnected: feed removed or disabled")
			delete(f.runners, id)
			delete(f.liveFeeds, id)
		}
	}

	for id, feed := range want {
		if _, running := f.runners[id]; running {
			old := f.liveFeeds[id]
			if old != nil && !quoteFeedNeedsRestart(*old.Load(), feed) {
				f.bindLiveFeed(id, feed)
				continue
			}
			f.stopRunnerLocked(id, f.runners[id])
			delete(f.runners, id)
		}
		f.startRunnerLocked(ctx, id, feed)
	}

	f.h.Log.Log(logger.TypeSys, logger.CodeOK, "quote feeds synced", "active", len(f.runners))
	return nil
}

func (f *Feeds) startRunnerLocked(ctx context.Context, id int64, feed model.QuoteFeed) {
	runCtx, cancel := context.WithCancel(ctx)
	feedPtr := f.bindLiveFeed(id, feed)
	conn := f.newConnector(feedPtr)
	if conn == nil {
		cancel()
		return
	}
	done := make(chan struct{})
	f.runners[id] = feedRunner{cancel: cancel, conn: conn, done: done}
	f.status.Journal(id, status.JournalInfo,
		fmt.Sprintf("connecting to %s (%s)", feed.Datafeed.FeedServer, feed.Datafeed.Module))
	// a FIX feed reports Connected on logon, not on process start
	if conn.Type() != provider.TypeFIX {
		f.status.Connected(feed.Datafeed.DatafeedID)
		f.status.Journal(id, status.JournalInfo, "connected")
	}
	f.h.Go(func() {
		defer close(done)
		if err := conn.Run(runCtx); err != nil && runCtx.Err() == nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "quote feed stopped",
				"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
			f.status.Disconnected(feed.Datafeed.DatafeedID)
			f.status.Journal(id, status.JournalErr, "feed stopped: "+err.Error())
		}
	})
}

func quoteFeedNeedsRestart(old, next model.QuoteFeed) bool {
	if old.Datafeed.Module != next.Datafeed.Module ||
		old.Datafeed.FeedServer != next.Datafeed.FeedServer ||
		old.Datafeed.FeedLogin != next.Datafeed.FeedLogin ||
		old.Datafeed.FeedPassword != next.Datafeed.FeedPassword {
		return true
	}
	if !reflect.DeepEqual(old.Params, next.Params) {
		return true
	}
	if !reflect.DeepEqual(old.Translates, next.Translates) {
		return true
	}
	return false
}

func (f *Feeds) reloadOne(ctx context.Context, datafeedID int64) error {
	if !f.ownsFeed(datafeedID) {
		f.mu.Lock()
		if runner, ok := f.runners[datafeedID]; ok {
			f.stopRunnerLocked(datafeedID, runner)
			delete(f.runners, datafeedID)
			delete(f.liveFeeds, datafeedID)
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
	defer f.mu.Unlock()

	if runner, ok := f.runners[datafeedID]; ok {
		f.stopRunnerLocked(datafeedID, runner)
		delete(f.runners, datafeedID)
	}

	if feed == nil || !feed.IsQuoteEnabled() || len(feed.Translates) == 0 {
		delete(f.liveFeeds, datafeedID)
		f.status.Disconnected(datafeedID)
		return nil
	}
	if _, ok := provider.ModuleType(feed.Datafeed.Module); !ok {
		delete(f.liveFeeds, datafeedID)
		f.status.Disconnected(datafeedID)
		return nil
	}

	f.startRunnerLocked(ctx, datafeedID, *feed)
	return nil
}

func (f *Feeds) stopAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for id, runner := range f.runners {
		f.stopRunnerLocked(id, runner)
		f.status.Disconnected(id)
		delete(f.runners, id)
		delete(f.liveFeeds, id)
	}
}

func (f *Feeds) newConnector(feedPtr *atomic.Pointer[model.QuoteFeed]) streamConnector {
	tickCh := make(chan provider.RawTick, 256)
	feed := feedPtr.Load()
	typ, _ := provider.ModuleType(feed.Datafeed.Module)

	switch typ {
	case provider.TypeFIX:
		settings, err := fixconfig.FromFeed(*feed)
		if err != nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "fix settings build failed",
				"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
			return nil
		}
		cfgPath, err := fixconfig.ResolveConfigPath(f.h.Cfg.Quote.FixConfigDir, *feed, settings)
		if err != nil {
			f.h.Log.Log(logger.TypeNet, logger.CodeErr, "fix config build failed",
				"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
			return nil
		}
		inner := fixquotes.NewConnector(settings, cfgPath, fixconfig.ExternalSymbols(*feed), f.h.Log, tickCh)
		id := feed.Datafeed.DatafeedID
		inner.OnSession = func(connected bool, detail string) {
			if connected {
				f.status.Connected(id)
				f.status.Journal(id, status.JournalInfo, detail)
				return
			}
			f.status.Disconnected(id)
			f.status.Journal(id, status.JournalWarn, detail)
		}
		return &fixConnector{
			inner: inner,
			feed:  feedPtr,
			ticks: tickCh,
			state: filter.New(),
			f:     f,
		}
	default:
		return nil
	}
}

// fixConnector wraps FIX and forwards ticks to the feed handler.
type fixConnector struct {
	inner *fixquotes.Connector
	feed  *atomic.Pointer[model.QuoteFeed]
	ticks <-chan provider.RawTick
	state *filter.State
	f     *Feeds
}

func (c *fixConnector) Type() provider.ConnectorType { return c.inner.Type() }

func (c *fixConnector) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.inner.Run(ctx)
	}()

	logons := c.inner.Logons()
	for {
		select {
		case <-ctx.Done():
			_ = c.inner.Close()
			select {
			case <-errCh:
			case <-time.After(runnerStopTimeout):
			}
			return ctx.Err()
		case err := <-errCh:
			return err
		case raw, ok := <-c.ticks:
			if !ok {
				return nil
			}
			// a reconnect is a break in the stream, so the tick that follows it cannot be filtered
			if n := c.inner.Logons(); n != logons {
				logons = n
				c.state.Reset()
			}
			c.f.handleRawTick(ctx, *c.feed.Load(), c.state, raw)
		}
	}
}

func (c *fixConnector) Close() error { return c.inner.Close() }

func (f *Feeds) handleRawTick(ctx context.Context, feed model.QuoteFeed, st *filter.State, raw provider.RawTick) {
	tick, ok := translate.ApplyMarkup(feed, raw)
	if !ok {
		return
	}

	set, hasSet := feed.Settings[tick.SymbolID]
	// every discard path feeds the raw series, that is what it is for
	discard := func() {
		if hasSet && set.CollectRaw() && f.influx != nil {
			f.influx.WriteRawTick(*tick)
		}
	}

	if hasSet && !set.RealtimeAllowed() {
		st.MarkBreak(tick.SymbolID)
		if st.WarnOnce(tick.SymbolID) {
			f.h.Log.Log(logger.TypeSys, logger.CodeWarn, "symbol drops ticks, realtime flag off",
				"datafeed_id", tick.DatafeedID, "symbol", tick.Symbol, "tick_flags", set.TickFlags)
		}
		discard()
		return
	}

	// a closed quote session is no stream at all, so the next tick starts a fresh channel
	if !f.isQuoteSessionOpen(feed, tick.SymbolID) {
		st.MarkBreak(tick.SymbolID)
		discard()
		return
	}

	if !st.Apply(set, tick) {
		discard()
		return
	}

	translate.ApplySpread(set, tick)

	if err := f.cache.Put(ctx, *tick); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "tick cache failed",
			"datafeed_id", tick.DatafeedID, "symbol", tick.Symbol, "error", err.Error())
	}

	if f.influx != nil {
		f.influx.WriteTick(*tick)
	}

	payload := f.publishTick(*tick)
	if f.status != nil && len(payload) > 0 {
		f.status.Tick(tick.DatafeedID, int64(len(payload)))
	}
}

func (f *Feeds) isQuoteSessionOpen(feed model.QuoteFeed, symbolID int64) bool {
	// only this symbol's rows, the feed carries every symbol's and the tick path runs hot
	var windows []session.Window
	for _, s := range feed.Sessions {
		if s.SymbolID != symbolID {
			continue
		}
		windows = append(windows, session.Window{
			SymbolID: s.SymbolID,
			Day:      s.Day,
			Open:     s.Open,
			Close:    s.Close,
		})
	}
	return session.IsQuoteOpen(symbolID, windows, time.Now().UTC())
}

func (f *Feeds) publishTick(tick model.Tick) []byte {
	payload, err := json.Marshal(tick)
	if err != nil {
		return nil
	}
	subject := model.SubjectTick(tick.Symbol)
	if err := f.h.Nats.NC.Publish(subject, payload); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "tick publish failed",
			"subject", subject, "error", err.Error())
	}
	return payload
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
			f.stopRunnerLocked(evt.DatafeedID, runner)
			delete(f.runners, evt.DatafeedID)
			delete(f.liveFeeds, evt.DatafeedID)
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
			f.stopRunnerLocked(snap.DatafeedID, runner)
			delete(f.runners, snap.DatafeedID)
			delete(f.liveFeeds, snap.DatafeedID)
		}
		f.mu.Unlock()
		f.status.Disconnected(snap.DatafeedID)
		return
	}

	feed := configclient.FromSnapshot(snap)
	f.mu.Lock()
	if _, running := f.runners[snap.DatafeedID]; running {
		if ptr := f.liveFeeds[snap.DatafeedID]; ptr != nil && !quoteFeedNeedsRestart(*ptr.Load(), feed) {
			f.bindLiveFeed(snap.DatafeedID, feed)
			f.mu.Unlock()
			return
		}
	}
	f.mu.Unlock()

	if err := f.reloadOne(context.Background(), snap.DatafeedID); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "config snapshot reload failed",
			"datafeed_id", snap.DatafeedID, "error", err.Error())
	}
}
