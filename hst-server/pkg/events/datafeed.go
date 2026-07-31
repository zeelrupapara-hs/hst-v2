package events

import (
	"encoding/json"
	"fmt"

	"hstserver/model"
	"hstserver/pkg/nats"
)

const (
	SubjectDatafeedCreated = "hstserver.datafeed.created"
	SubjectDatafeedUpdated = "hstserver.datafeed.updated"
	SubjectDatafeedDeleted = "hstserver.datafeed.deleted"
	SubjectDatafeedConfig  = "hstserver.datafeed.config"
)

// DatafeedEvent notifies workers that feed configuration changed.
type DatafeedEvent struct {
	DatafeedID int64 `json:"datafeed_id"`
	Mode       int32 `json:"mode"`
	Enable     int16 `json:"enable"`
}

// WorkerParam is a key/value feed parameter for worker consumption.
type WorkerParam struct {
	ParamKey string `json:"param_key"`
	Value    string `json:"value"`
}

// WorkerTranslate is symbol mapping for quote workers.
type WorkerTranslate struct {
	TranslateID int64  `json:"translate_id"`
	SymbolID    int64  `json:"symbol_id"`
	Symbol      string `json:"symbol"`
	Source      string `json:"source"`
	BidMarkup   int32  `json:"bid_markup"`
	AskMarkup   int32  `json:"ask_markup"`
	Digits      int16  `json:"digits"`
}

// WorkerDatafeedConfig is the full config snapshot workers need without DB access.
type WorkerDatafeedConfig struct {
	DatafeedID       int64                `json:"datafeed_id"`
	Name             string               `json:"name"`
	Module           string               `json:"module"`
	Enable           model.DatafeedEnable `json:"enable"`
	Mode             model.FeederFlags    `json:"mode"`
	FeedServer       string               `json:"feed_server"`
	FeedLogin        int64                `json:"feed_login"`
	FeedPassword     string               `json:"feed_password"`
	TimeoutReconnect int32                `json:"timeout_reconnect"`
	Params           []WorkerParam        `json:"params"`
	Translates       []WorkerTranslate    `json:"translates"`
}

// ConfigSubject returns the per-feed config subject.
func ConfigSubject(datafeedID int64) string {
	return fmt.Sprintf("%s.%d", SubjectDatafeedConfig, datafeedID)
}

// PublishDatafeed sends a datafeed lifecycle event on NATS.
func PublishDatafeed(nc *nats.Nats, subject string, datafeedID int64, mode model.FeederFlags, enable model.DatafeedEnable) error {
	if nc == nil || nc.NC == nil {
		return nil
	}

	evt := DatafeedEvent{
		DatafeedID: datafeedID,
		Mode:       int32(mode),
		Enable:     int16(enable),
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal datafeed event: %w", err)
	}
	if err := nc.NC.Publish(subject, payload); err != nil {
		return fmt.Errorf("publish datafeed event: %w", err)
	}
	return nil
}

// PublishConfigSnapshot sends the full worker config for one datafeed.
func PublishConfigSnapshot(nc *nats.Nats, cfg WorkerDatafeedConfig) error {
	if nc == nil || nc.NC == nil {
		return nil
	}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal datafeed config: %w", err)
	}
	if err := nc.NC.Publish(ConfigSubject(cfg.DatafeedID), payload); err != nil {
		return fmt.Errorf("publish datafeed config: %w", err)
	}
	return nil
}
