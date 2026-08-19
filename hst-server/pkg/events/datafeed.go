package events

import (
	"encoding/json"
	"fmt"

	"hstserver/model"
	"hstserver/pkg/nats"
)

const (
	SubjectDatafeedCreated = "system.datafeeds.created"
	SubjectDatafeedUpdated = "system.datafeeds.updated"
	SubjectDatafeedDeleted = "system.datafeeds.deleted"
	SubjectDatafeedConfig  = "system.datafeeds.config"
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

// WorkerSymbolSession is one quote session window for a symbol weekday.
type WorkerSymbolSession struct {
	SymbolID int64 `json:"symbol_id"`
	Day      int16 `json:"day"`
	Open     int32 `json:"open"`
	Close    int32 `json:"close"`
}

// WorkerSymbolSettings is per-symbol quote handling config for quote workers.
type WorkerSymbolSettings struct {
	SymbolID        int64   `json:"symbol_id"`
	Digits          int16   `json:"digits"`
	Point           float64 `json:"point"`
	TickFlags       int32   `json:"tick_flags"`
	TickBookDepth   int32   `json:"tick_book_depth"`
	CalcMode        int16   `json:"calc_mode"`
	TickChartMode   int16   `json:"tick_chart_mode"`
	SpliceType      int16   `json:"splice_type"`
	FilterSoft      int32   `json:"filter_soft"`
	FilterSoftTicks int32   `json:"filter_soft_ticks"`
	FilterHard      int32   `json:"filter_hard"`
	FilterHardTicks int32   `json:"filter_hard_ticks"`
	FilterDiscard   int32   `json:"filter_discard"`
	FilterSpreadMin int32   `json:"filter_spread_min"`
	FilterSpreadMax int32   `json:"filter_spread_max"`
	FilterGap       int32   `json:"filter_gap"`
	FilterGapTicks  int32   `json:"filter_gap_ticks"`
	Spread          int32   `json:"spread"`
	SpreadBalance   int32   `json:"spread_balance"`
}

// WorkerDatafeedConfig is the full config snapshot workers need without DB access.
type WorkerDatafeedConfig struct {
	DatafeedID       int64                  `json:"datafeed_id"`
	Name             string                 `json:"name"`
	Module           string                 `json:"module"`
	Enable           model.DatafeedEnable   `json:"enable"`
	FeedIndex        int32                  `json:"feed_index"`
	Mode             model.FeederFlags      `json:"mode"`
	FeedServer       string                 `json:"feed_server"`
	FeedLogin        string                 `json:"feed_login"`
	FeedPassword     string                 `json:"feed_password"`
	TimeoutReconnect int32                  `json:"timeout_reconnect"`
	Params           []WorkerParam          `json:"params"`
	Translates       []WorkerTranslate      `json:"translates"`
	Sessions         []WorkerSymbolSession  `json:"sessions"`
	Settings         []WorkerSymbolSettings `json:"settings"`
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
