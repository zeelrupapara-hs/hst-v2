package status

import (
	"hstquote/model"

	"encoding/json"
	"time"

	"hstquote/pkg/nats"
)

// Event is runtime telemetry for hst-server aggregation.
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

// JournalEvent is one line of a feed's operating journal.
type JournalEvent struct {
	DatafeedID int64  `json:"datafeed_id"`
	Code       int32  `json:"code"`
	Message    string `json:"message"`
	Time       int64  `json:"time"`
}

// Journal codes follow the platform logger: 0 info, 1 warning, 2 error.
const (
	JournalInfo int32 = 0
	JournalWarn int32 = 1
	JournalErr  int32 = 2
)

// Publisher sends quote worker status to NATS.
type Publisher struct {
	nc         *nats.Nats
	instanceID string
}

func New(nc *nats.Nats, instanceID string) *Publisher {
	if instanceID == "" {
		instanceID = "hstquote"
	}
	return &Publisher{nc: nc, instanceID: instanceID}
}

func (p *Publisher) Publish(evt Event) {
	if p == nil || p.nc == nil || p.nc.NC == nil {
		return
	}
	evt.Service = "hstquote"
	if evt.InstanceID == "" {
		evt.InstanceID = p.instanceID
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return
	}
	_ = p.nc.NC.Publish(model.SubjectStatus(evt.DatafeedID), payload)
}

func (p *Publisher) Connected(datafeedID int64) {
	p.Publish(Event{
		DatafeedID:  datafeedID,
		Connected:   true,
		SysLastTime: time.Now().UTC().UnixNano(),
	})
}

func (p *Publisher) Disconnected(datafeedID int64) {
	p.Publish(Event{
		DatafeedID: datafeedID,
		Connected:  false,
	})
}

// Book counts one Market Depth change; no feed produces books yet, so it waits for one that does.
func (p *Publisher) Book(datafeedID int64, bytes int64) {
	p.Publish(Event{
		DatafeedID:         datafeedID,
		Connected:          true,
		SysLastTime:        time.Now().UTC().UnixNano(),
		BooksDelta:         1,
		BytesReceivedDelta: bytes,
	})
}

// Journal announces one feed journal line for the admin to keep.
func (p *Publisher) Journal(datafeedID int64, code int32, message string) {
	if p == nil || p.nc == nil || p.nc.NC == nil {
		return
	}
	payload, err := json.Marshal(JournalEvent{
		DatafeedID: datafeedID,
		Code:       code,
		Message:    message,
		Time:       time.Now().UTC().UnixNano(),
	})
	if err != nil {
		return
	}
	_ = p.nc.NC.Publish(model.SubjectJournal(datafeedID), payload)
}

func (p *Publisher) Tick(datafeedID int64, bytes int64) {
	p.Publish(Event{
		DatafeedID:         datafeedID,
		Connected:          true,
		SysLastTime:        time.Now().UTC().UnixNano(),
		TicksDelta:         1,
		BytesReceivedDelta: bytes,
	})
}
