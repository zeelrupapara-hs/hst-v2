package workerstatus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/db"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"

	natscore "github.com/nats-io/nats.go"
)

const (
	subjectQuoteStatus = "hstquote.status.>"
	subjectNewsStatus  = "hstnews.status.>"
	groupQuoteStatus   = "hstserver-quote-status"
	groupNewsStatus    = "hstserver-news-status"

	wsThrottle  = time.Second
	flushPeriod = 5 * time.Second
)

// Event is runtime telemetry published by ingestion workers.
type Event struct {
	DatafeedID         int64  `json:"datafeed_id"`
	Service            string `json:"service"`
	InstanceID         string `json:"instance_id"`
	Connected          bool   `json:"connected"`
	SysLastTime        int64  `json:"sys_last_time"`
	TicksDelta         int64  `json:"ticks_delta"`
	NewsDelta          int64  `json:"news_delta"`
	BytesReceivedDelta int64  `json:"bytes_received_delta"`
}

// Notifier publishes admin websocket events (wired from v1.HttpServer.NotifyWS).
type Notifier func(subject, event string, payload any)

type feedState struct {
	baseTicks    int64
	baseNews     int64
	baseBytes    int64
	pendingTicks int64
	pendingNews  int64
	pendingBytes int64
	sysLastTime  int64
	lastNotify   time.Time
}

// Subscriber applies worker telemetry to hst.datafeeds and pushes live updates to admin.
type Subscriber struct {
	db       *db.PostgresDB
	log      *logger.Logger
	subs     []*natscore.Subscription
	notify   Notifier
	feeds    map[int64]*feedState
	mu       sync.Mutex
	stopCh   chan struct{}
	stopOnce sync.Once
}

func New(database *db.PostgresDB, log *logger.Logger) *Subscriber {
	return &Subscriber{
		db:     database,
		log:    log,
		feeds:  make(map[int64]*feedState),
		stopCh: make(chan struct{}),
	}
}

func (s *Subscriber) SetNotifier(n Notifier) {
	s.mu.Lock()
	s.notify = n
	s.mu.Unlock()
}

func (s *Subscriber) Start(nc *nats.Nats) error {
	if nc == nil || nc.NC == nil {
		return nil
	}
	if err := s.loadBaselines(context.Background()); err != nil {
		s.log.Log(logger.TypeNet, logger.CodeWarn, "worker status baseline load failed", "error", err.Error())
	}
	quoteSub, err := nc.NC.QueueSubscribe(subjectQuoteStatus, groupQuoteStatus, s.onQuoteStatus)
	if err != nil {
		return err
	}
	newsSub, err := nc.NC.QueueSubscribe(subjectNewsStatus, groupNewsStatus, s.onNewsStatus)
	if err != nil {
		_ = quoteSub.Unsubscribe()
		return err
	}
	s.mu.Lock()
	s.subs = []*natscore.Subscription{quoteSub, newsSub}
	s.mu.Unlock()
	go s.flushLoop()
	s.log.Log(logger.TypeNet, logger.CodeOK, "worker status subscriber started")
	return nil
}

func (s *Subscriber) Stop() {
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
	s.subs = nil
	s.flushAll(context.Background())
}

func (s *Subscriber) loadBaselines(ctx context.Context) error {
	rows, err := s.db.DB.Query(ctx,
		`SELECT datafeed_id, ticks_count, news_count, bytes_received FROM hst.datafeeds`)
	if err != nil {
		return err
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	for rows.Next() {
		var id, ticks, news, bytes int64
		if err := rows.Scan(&id, &ticks, &news, &bytes); err != nil {
			return err
		}
		st := s.feed(id)
		st.baseTicks = ticks
		st.baseNews = news
		st.baseBytes = bytes
	}
	return rows.Err()
}

func (s *Subscriber) onQuoteStatus(msg *natscore.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.TicksDelta != 0 || evt.BytesReceivedDelta != 0 {
		s.applyStats(evt, evt.TicksDelta, 0, evt.BytesReceivedDelta)
		return
	}
	s.applyConnection(evt)
}

func (s *Subscriber) onNewsStatus(msg *natscore.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.NewsDelta != 0 || evt.BytesReceivedDelta != 0 {
		s.applyStats(evt, 0, evt.NewsDelta, evt.BytesReceivedDelta)
		return
	}
	s.applyConnection(evt)
}

func (s *Subscriber) applyConnection(evt Event) {
	conn := model.DatafeedSysConnection_down
	if evt.Connected {
		conn = model.DatafeedSysConnection_connected
	}
	_, err := s.db.DB.Exec(context.Background(),
		`UPDATE hst.datafeeds SET
		    sys_connection = $2,
		    sys_last_time = CASE WHEN $3::bigint > 0 THEN $3::bigint ELSE sys_last_time END
		  WHERE datafeed_id = $1`,
		evt.DatafeedID, conn, evt.SysLastTime)
	if err != nil {
		s.log.Log(logger.TypeNet, logger.CodeWarn, "quote status update failed",
			"datafeed_id", evt.DatafeedID, "error", err.Error())
		return
	}
	if evt.SysLastTime > 0 {
		s.mu.Lock()
		s.feed(evt.DatafeedID).sysLastTime = evt.SysLastTime
		s.mu.Unlock()
	}
	s.notifyRuntime(model.EventDatafeedStatusUpdated, model.DatafeedRuntime{
		DatafeedID:    evt.DatafeedID,
		SysConnection: &conn,
		SysLastTime:   evt.SysLastTime,
	})
}

func (s *Subscriber) applyStats(evt Event, ticks, news, bytes int64) {
	s.mu.Lock()
	st := s.feed(evt.DatafeedID)
	st.pendingTicks += ticks
	st.pendingNews += news
	st.pendingBytes += bytes
	if evt.SysLastTime > 0 {
		st.sysLastTime = evt.SysLastTime
	}
	now := time.Now()
	throttled := !st.lastNotify.IsZero() && now.Sub(st.lastNotify) < wsThrottle
	runtime := s.runtimeLocked(evt.DatafeedID, st)
	if !throttled {
		st.lastNotify = now
	}
	s.mu.Unlock()

	if !throttled {
		s.notifyRuntime(model.EventDatafeedStatsUpdated, runtime)
	}
}

func (s *Subscriber) feed(id int64) *feedState {
	st, ok := s.feeds[id]
	if !ok {
		st = &feedState{}
		s.feeds[id] = st
	}
	return st
}

func (s *Subscriber) runtimeLocked(id int64, st *feedState) model.DatafeedRuntime {
	return model.DatafeedRuntime{
		DatafeedID:    id,
		TicksCount:    st.baseTicks + st.pendingTicks,
		NewsCount:     st.baseNews + st.pendingNews,
		BytesReceived: st.baseBytes + st.pendingBytes,
		SysLastTime:   st.sysLastTime,
	}
}

func (s *Subscriber) notifyRuntime(event string, payload model.DatafeedRuntime) {
	s.mu.Lock()
	notify := s.notify
	s.mu.Unlock()
	if notify == nil {
		return
	}
	notify(model.SubjectDatafeed, event, payload)
}

func (s *Subscriber) flushLoop() {
	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.flushAll(context.Background())
		}
	}
}

func (s *Subscriber) flushAll(ctx context.Context) {
	type pending struct {
		id    int64
		ticks int64
		news  int64
		bytes int64
		last  int64
	}
	var batch []pending

	s.mu.Lock()
	for id, st := range s.feeds {
		if st.pendingTicks == 0 && st.pendingNews == 0 && st.pendingBytes == 0 {
			continue
		}
		batch = append(batch, pending{
			id:    id,
			ticks: st.pendingTicks,
			news:  st.pendingNews,
			bytes: st.pendingBytes,
			last:  st.sysLastTime,
		})
		st.baseTicks += st.pendingTicks
		st.baseNews += st.pendingNews
		st.baseBytes += st.pendingBytes
		st.pendingTicks = 0
		st.pendingNews = 0
		st.pendingBytes = 0
	}
	s.mu.Unlock()

	for _, p := range batch {
		_, err := s.db.DB.Exec(ctx,
			`UPDATE hst.datafeeds SET
			    ticks_count = ticks_count + $2,
			    news_count = news_count + $3,
			    bytes_received = bytes_received + $4,
			    sys_last_time = CASE WHEN $5::bigint > 0 THEN $5::bigint ELSE sys_last_time END
			  WHERE datafeed_id = $1`,
			p.id, p.ticks, p.news, p.bytes, p.last)
		if err != nil {
			s.log.Log(logger.TypeNet, logger.CodeWarn, "worker stats flush failed",
				"datafeed_id", p.id, "error", err.Error())
		}
	}
}
