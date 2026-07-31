package status

import (
	"encoding/json"
	"fmt"
	"time"

	"hstnews/pkg/nats"
)

const subjectTemplate = "hstnews.status.%d"

// Event is runtime telemetry for hst-server aggregation.
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

// Publisher sends news worker status to NATS.
type Publisher struct {
	nc         *nats.Nats
	instanceID string
}

func New(nc *nats.Nats, instanceID string) *Publisher {
	if instanceID == "" {
		instanceID = "hstnews"
	}
	return &Publisher{nc: nc, instanceID: instanceID}
}

func (p *Publisher) Publish(evt Event) {
	if p == nil || p.nc == nil || p.nc.NC == nil {
		return
	}
	evt.Service = "hstnews"
	if evt.InstanceID == "" {
		evt.InstanceID = p.instanceID
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return
	}
	_ = p.nc.NC.Publish(fmt.Sprintf(subjectTemplate, evt.DatafeedID), payload)
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

func (p *Publisher) News(datafeedID int64, count int64, bytes int64) {
	p.Publish(Event{
		DatafeedID:         datafeedID,
		Connected:          true,
		SysLastTime:        time.Now().UTC().UnixNano(),
		NewsDelta:          count,
		BytesReceivedDelta: bytes,
	})
}
