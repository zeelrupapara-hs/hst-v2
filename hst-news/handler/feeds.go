package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"hstnews/connector"
	"hstnews/internal/configclient"
	newscache "hstnews/internal/newscache"
	"hstnews/internal/normalize"
	"hstnews/internal/status"
	"hstnews/model"
	"hstnews/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

type feedRunner struct {
	cancel context.CancelFunc
	feed   model.NewsFeed
}

// Feeds manages polling loops for configured news sources.
type Feeds struct {
	h      *Handler
	config *configclient.Client
	cache  *newscache.Store
	status *status.Publisher

	mu      sync.Mutex
	runners map[int64]feedRunner
}

func newFeeds(h *Handler) *Feeds {
	return &Feeds{
		h:       h,
		config:  configclient.New(h.Cfg.News.ServerURL, h.Cfg.News.ServiceToken),
		cache:   newscache.New(h.Redis),
		status:  status.New(h.Nats, h.Cfg.Nats.Name),
		runners: make(map[int64]feedRunner),
	}
}

func (f *Feeds) ownsFeed(datafeedID int64) bool {
	count := f.h.Cfg.News.InstanceCount
	if count <= 1 {
		return true
	}
	idx := f.h.Cfg.News.InstanceIndex
	return datafeedID%int64(count) == int64(idx)
}

func (f *Feeds) reload(ctx context.Context) error {
	feeds, err := f.config.ListNewsFeeds(ctx)
	if err != nil {
		return err
	}

	want := make(map[int64]model.NewsFeed, len(feeds))
	for _, feed := range feeds {
		if !feed.IsNewsEnabled() {
			continue
		}
		if !f.ownsFeed(feed.Datafeed.DatafeedID) {
			continue
		}
		if _, ok := connector.ForModule(feed.Datafeed.Module); !ok {
			f.h.Log.Log(logger.TypeSys, logger.CodeWarn, "unsupported news module, skipping",
				"datafeed_id", feed.Datafeed.DatafeedID, "module", feed.Datafeed.Module)
			continue
		}
		want[feed.Datafeed.DatafeedID] = feed
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for id, runner := range f.runners {
		if _, keep := want[id]; !keep {
			runner.cancel()
			f.status.Disconnected(id)
			delete(f.runners, id)
		}
	}

	for id, feed := range want {
		if runner, running := f.runners[id]; running {
			if reflect.DeepEqual(runner.feed, feed) {
				continue
			}
			// config changed: restart the poller with the new feed
			runner.cancel()
			delete(f.runners, id)
		}
		runCtx, cancel := context.WithCancel(ctx)
		f.runners[id] = feedRunner{cancel: cancel, feed: feed}
		feedCopy := feed
		f.status.Connected(feedCopy.Datafeed.DatafeedID)
		f.h.Go(func() { f.runFeed(runCtx, feedCopy) })
	}

	f.h.Log.Log(logger.TypeSys, logger.CodeOK, "news feeds synced", "active", len(f.runners))
	return nil
}

func (f *Feeds) reloadOne(ctx context.Context, datafeedID int64) error {
	if !f.ownsFeed(datafeedID) {
		f.mu.Lock()
		if runner, ok := f.runners[datafeedID]; ok {
			runner.cancel()
			delete(f.runners, datafeedID)
		}
		f.mu.Unlock()
		f.status.Disconnected(datafeedID)
		return nil
	}

	feed, err := f.config.GetNewsFeed(ctx, datafeedID)
	if err != nil {
		return err
	}

	f.mu.Lock()
	if runner, ok := f.runners[datafeedID]; ok {
		runner.cancel()
		delete(f.runners, datafeedID)
	}
	f.mu.Unlock()
	f.status.Disconnected(datafeedID)

	if feed == nil || !feed.IsNewsEnabled() {
		return nil
	}
	if _, ok := connector.ForModule(feed.Datafeed.Module); !ok {
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	f.mu.Lock()
	f.runners[datafeedID] = feedRunner{cancel: cancel, feed: *feed}
	f.mu.Unlock()
	f.status.Connected(datafeedID)

	feedCopy := *feed
	f.h.Go(func() { f.runFeed(runCtx, feedCopy) })
	return nil
}

func (f *Feeds) stopAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for id, runner := range f.runners {
		runner.cancel()
		f.status.Disconnected(id)
		delete(f.runners, id)
	}
}

func (f *Feeds) runFeed(ctx context.Context, feed model.NewsFeed) {
	interval := normalize.PollInterval(feed)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	f.h.Log.Log(logger.TypeSys, logger.CodeOK, "news feed started",
		"datafeed_id", feed.Datafeed.DatafeedID, "module", feed.Datafeed.Module, "interval", interval.String())

	f.pollOnce(ctx, feed)

	for {
		select {
		case <-ctx.Done():
			f.status.Disconnected(feed.Datafeed.DatafeedID)
			f.h.Log.Log(logger.TypeSys, logger.CodeOK, "news feed stopped",
				"datafeed_id", feed.Datafeed.DatafeedID)
			return
		case <-ticker.C:
			f.pollOnce(ctx, feed)
		}
	}
}

func (f *Feeds) pollOnce(ctx context.Context, feed model.NewsFeed) {
	conn, ok := connector.ForModule(feed.Datafeed.Module)
	if !ok {
		return
	}

	raw, bytes, err := conn.Fetch(ctx, feed)
	if err != nil {
		f.status.Disconnected(feed.Datafeed.DatafeedID)
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news fetch failed",
			"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
		return
	}

	now := time.Now().UTC()
	items := normalize.Items(feed, raw, now)
	newItems, err := f.cache.FilterNew(ctx, items)
	if err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news dedup failed",
			"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
		return
	}

	// every successful poll is a heartbeat; hst-server only counts deltas
	f.status.Publish(status.Event{
		DatafeedID:         feed.Datafeed.DatafeedID,
		Connected:          true,
		SysLastTime:        now.UnixNano(),
		NewsDelta:          int64(len(newItems)),
		BytesReceivedDelta: bytes,
	})
	if len(newItems) == 0 {
		return
	}

	if err := f.cache.MarkSeen(ctx, newItems); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news cache mark failed",
			"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
		return
	}
	if err := f.cache.PrependFeed(ctx, newItems); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news cache store failed",
			"datafeed_id", feed.Datafeed.DatafeedID, "error", err.Error())
	}

	for _, item := range newItems {
		f.publishItem(item)
	}

	f.h.Log.Log(logger.TypeNotify, logger.CodeOK, "news items ingested",
		"datafeed_id", feed.Datafeed.DatafeedID, "count", len(newItems))
}

func (f *Feeds) publishItem(item model.NewsItem) {
	payload, err := json.Marshal(item)
	if err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news marshal failed",
			"datafeed_id", item.DatafeedID, "error", err.Error())
		return
	}

	subject := fmt.Sprintf(SubjectNewsItem, item.DatafeedID)
	if err := f.h.Nats.NC.Publish(subject, payload); err != nil {
		f.h.Log.Log(logger.TypeNet, logger.CodeErr, "news publish failed",
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

	if msg.Subject == SubjectDatafeedDeleted || !evt.HasNewsFlag() {
		f.mu.Lock()
		if runner, ok := f.runners[evt.DatafeedID]; ok {
			runner.cancel()
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
	if snap.Enable != model.DatafeedEnable_enabled || (snap.Mode&model.FeederFlags_news) == 0 {
		f.mu.Lock()
		if runner, ok := f.runners[snap.DatafeedID]; ok {
			runner.cancel()
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
