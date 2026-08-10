package workerstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/db"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/pkg/nats"

	natscore "github.com/nats-io/nats.go"
)

const (
	subjectQuoteStatus  = "hstquote.status.>"
	subjectNewsStatus   = "hstnews.status.>"
	subjectQuoteJournal = "hstquote.journal.>"
	groupQuoteStatus    = "hstserver-quote-status"
	groupNewsStatus     = "hstserver-news-status"
	groupQuoteJournal   = "hstserver-quote-journal"

	wsThrottle   = time.Second
	flushPeriod  = 5 * time.Second
	statusPeriod = 10 * time.Second
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
	BooksDelta         int64  `json:"books_delta"`
	BytesReceivedDelta int64  `json:"bytes_received_delta"`
}

// JournalEvent is one line of a feed's operating journal, kept in hst.journal under the
// feed's own channel so the panel can show it beside the feed.
type JournalEvent struct {
	DatafeedID int64  `json:"datafeed_id"`
	Code       int32  `json:"code"`
	Message    string `json:"message"`
	Time       int64  `json:"time"`
}

// Notifier publishes admin websocket events (wired from v1.HttpServer.NotifyWS).
type Notifier func(subject, event string, payload any)

type feedState struct {
	baseTicks    int64
	baseNews     int64
	baseBooks    int64
	baseBytes    int64
	pendingTicks int64
	pendingNews  int64
	pendingBooks int64
	pendingBytes int64
	sysLastTime  int64
	sysConn      *model.DatafeedSysConnection
	lastNotify   time.Time
}

// Subscriber applies worker telemetry to hst.datafeeds and pushes live updates to admin.
type Subscriber struct {
	db       *db.PostgresDB
	log      *logger.Logger
	journal  *journal.Journal
	subs     []*natscore.Subscription
	notify   Notifier
	feeds    map[int64]*feedState
	mu       sync.Mutex
	stopCh   chan struct{}
	stopOnce sync.Once
}

func New(database *db.PostgresDB, log *logger.Logger, jrn *journal.Journal) *Subscriber {
	return &Subscriber{
		db:      database,
		log:     log,
		journal: jrn,
		feeds:   make(map[int64]*feedState),
		stopCh:  make(chan struct{}),
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
	journalSub, err := nc.NC.QueueSubscribe(subjectQuoteJournal, groupQuoteJournal, s.onQuoteJournal)
	if err != nil {
		_ = quoteSub.Unsubscribe()
		_ = newsSub.Unsubscribe()
		return err
	}
	s.mu.Lock()
	s.subs = []*natscore.Subscription{quoteSub, newsSub, journalSub}
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
		`SELECT datafeed_id, ticks_count, news_count, books_count, bytes_received FROM hst.datafeeds`)
	if err != nil {
		return err
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	for rows.Next() {
		var id, ticks, news, books, bytes int64
		if err := rows.Scan(&id, &ticks, &news, &books, &bytes); err != nil {
			return err
		}
		st := s.feed(id)
		st.baseTicks = ticks
		st.baseNews = news
		st.baseBooks = books
		st.baseBytes = bytes
	}
	return rows.Err()
}

func (s *Subscriber) onQuoteStatus(msg *natscore.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.TicksDelta != 0 || evt.BooksDelta != 0 || evt.BytesReceivedDelta != 0 {
		s.applyStats(evt, evt.TicksDelta, 0, evt.BooksDelta, evt.BytesReceivedDelta)
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
		s.applyStats(evt, 0, evt.NewsDelta, 0, evt.BytesReceivedDelta)
		return
	}
	s.applyConnection(evt)
}

// onQuoteJournal keeps one feed journal line under the feed's channel; the Journal tab reads it back.
func (s *Subscriber) onQuoteJournal(msg *natscore.Msg) {
	var evt JournalEvent
	if err := json.Unmarshal(msg.Data, &evt); err != nil || evt.DatafeedID == 0 {
		return
	}
	at := evt.Time
	if at == 0 {
		at = time.Now().UTC().UnixNano()
	}
	_, err := s.db.DB.Exec(context.Background(),
		`INSERT INTO hst.journal (created_at, type, code, login, ip, channel, os, message, detail)
		 VALUES ($1, $2, $3, 0, NULL, $4, '', $5, '{}'::jsonb)`,
		at, int32(model.JournalType_datafeeds), evt.Code, fmt.Sprintf("datafeed:%d", evt.DatafeedID), evt.Message)
	if err != nil {
		s.log.Log(logger.TypeNet, logger.CodeWarn, "feed journal write failed",
			"datafeed_id", evt.DatafeedID, "error", err.Error())
	}
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
	s.mu.Lock()
	st := s.feed(evt.DatafeedID)
	prev := st.sysConn
	st.sysConn = &conn
	if evt.SysLastTime > 0 {
		st.sysLastTime = evt.SysLastTime
	}
	s.mu.Unlock()

	// only a flip is worth a journal line; the first status after boot counts when it is a connect
	if (prev == nil && evt.Connected) || (prev != nil && *prev != conn) {
		s.journalConnection(evt.DatafeedID, evt.Connected)
	}
	s.notifyRuntime(model.EventDatafeedStatusUpdated, model.DatafeedRuntime{
		DatafeedID:    evt.DatafeedID,
		SysConnection: &conn,
		SysLastTime:   evt.SysLastTime,
	})
}

// journalConnection writes the connect or disconnect transition to the server journal.
func (s *Subscriber) journalConnection(id int64, connected bool) {
	if s.journal == nil {
		return
	}

	name := fmt.Sprintf("datafeed #%d", id)
	var dbName string
	if err := s.db.DB.QueryRow(context.Background(),
		`SELECT name FROM hst.datafeeds WHERE datafeed_id = $1`, id).Scan(&dbName); err == nil {
		name = dbName
	}

	code, message := int32(logger.CodeOK), journal.DatafeedConnectedMsg(name)
	if !connected {
		code, message = int32(logger.CodeWarn), journal.DatafeedDisconnectedMsg(name)
	}

	if err := s.journal.Entry(context.Background(), &model.Journal{
		Type:    int32(model.JournalType_datafeeds),
		Code:    code,
		Channel: "system",
		Message: message,
	}); err != nil {
		s.log.Log(logger.TypeNet, logger.CodeWarn, "feed connection journal write failed",
			"datafeed_id", id, "error", err.Error())
	}
}

func (s *Subscriber) applyStats(evt Event, ticks, news, books, bytes int64) {
	s.mu.Lock()
	st := s.feed(evt.DatafeedID)
	st.pendingTicks += ticks
	st.pendingNews += news
	st.pendingBooks += books
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
		BooksCount:    st.baseBooks + st.pendingBooks,
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
	status := time.NewTicker(statusPeriod)
	defer status.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.flushAll(context.Background())
		case <-status.C:
			s.publishAllStatus()
		}
	}
}

// publishAllStatus pushes every feed's runtime on a fixed beat so the panel never goes stale.
func (s *Subscriber) publishAllStatus() {
	s.mu.Lock()
	batch := make([]model.DatafeedRuntime, 0, len(s.feeds))
	for id, st := range s.feeds {
		rt := s.runtimeLocked(id, st)
		rt.SysConnection = st.sysConn
		batch = append(batch, rt)
	}
	s.mu.Unlock()

	for _, rt := range batch {
		s.notifyRuntime(model.EventDatafeedStatusUpdated, rt)
	}
}

func (s *Subscriber) flushAll(ctx context.Context) {
	type pending struct {
		id    int64
		ticks int64
		news  int64
		books int64
		bytes int64
		last  int64
	}
	var batch []pending

	s.mu.Lock()
	for id, st := range s.feeds {
		if st.pendingTicks == 0 && st.pendingNews == 0 && st.pendingBooks == 0 && st.pendingBytes == 0 {
			continue
		}
		batch = append(batch, pending{
			id:    id,
			ticks: st.pendingTicks,
			news:  st.pendingNews,
			books: st.pendingBooks,
			bytes: st.pendingBytes,
			last:  st.sysLastTime,
		})
		st.baseTicks += st.pendingTicks
		st.baseNews += st.pendingNews
		st.baseBooks += st.pendingBooks
		st.baseBytes += st.pendingBytes
		st.pendingTicks = 0
		st.pendingNews = 0
		st.pendingBooks = 0
		st.pendingBytes = 0
	}
	s.mu.Unlock()

	for _, p := range batch {
		_, err := s.db.DB.Exec(ctx,
			`UPDATE hst.datafeeds SET
			    ticks_count = ticks_count + $2,
			    news_count = news_count + $3,
			    books_count = books_count + $4,
			    bytes_received = bytes_received + $5,
			    sys_last_time = CASE WHEN $6::bigint > 0 THEN $6::bigint ELSE sys_last_time END
			  WHERE datafeed_id = $1`,
			p.id, p.ticks, p.news, p.books, p.bytes, p.last)
		if err != nil {
			s.log.Log(logger.TypeNet, logger.CodeWarn, "worker stats flush failed",
				"datafeed_id", p.id, "error", err.Error())
		}
	}
}
