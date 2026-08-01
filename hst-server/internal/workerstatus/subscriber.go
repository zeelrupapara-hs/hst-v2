package workerstatus

import (
	"context"
	"encoding/json"
	"sync"

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

// Subscriber applies low-frequency worker connection state to hst.datafeeds.
// Session counters (ticks, bytes, news) are not persisted yet; live stats will
// use in-memory aggregation + WebSocket when implemented.
type Subscriber struct {
	db   *db.PostgresDB
	log  *logger.Logger
	subs []*natscore.Subscription
	mu   sync.Mutex
}

func New(database *db.PostgresDB, log *logger.Logger) *Subscriber {
	return &Subscriber{db: database, log: log}
}

func (s *Subscriber) Start(nc *nats.Nats) error {
	if nc == nil || nc.NC == nil {
		return nil
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
	s.log.Log(logger.TypeNet, logger.CodeOK, "worker status subscriber started")
	return nil
}

func (s *Subscriber) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
	s.subs = nil
}

func (s *Subscriber) onQuoteStatus(msg *natscore.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.TicksDelta != 0 || evt.BytesReceivedDelta != 0 {
		return
	}
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
	}
}

func (s *Subscriber) onNewsStatus(msg *natscore.Msg) {
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return
	}
	if evt.NewsDelta != 0 || evt.BytesReceivedDelta != 0 {
		return
	}
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
		s.log.Log(logger.TypeNet, logger.CodeWarn, "news status update failed",
			"datafeed_id", evt.DatafeedID, "error", err.Error())
	}
}
