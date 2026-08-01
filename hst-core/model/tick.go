package model

import (
	"encoding/json"
	"strconv"
	"time"
)

// Tick is one quote as the feed publishes it.
type Tick struct {
	DatafeedId int64   `json:"datafeed_id"`
	SymbolId   int64   `json:"symbol_id"`
	Symbol     string  `json:"symbol"`
	Source     string  `json:"source"`
	Digits     int32   `json:"digits"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	Last       float64 `json:"last"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Open       float64 `json:"open"`
	Close      float64 `json:"close"`
	Volume     float64 `json:"volume"`
	Time       int64   `json:"-"`
}

// the feed sends time as RFC3339; older producers send epoch nanoseconds
func (t *Tick) UnmarshalJSON(b []byte) error {
	type wire Tick

	var w struct {
		wire
		Time json.RawMessage `json:"time"`
	}
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}

	*t = Tick(w.wire)

	raw := string(w.Time)
	if raw == "" || raw == "null" {
		return nil
	}

	// a quoted value is a timestamp; a bare number is already epoch nanoseconds, and is read
	// as an integer so the low digits survive
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(w.Time, &text); err == nil {
			if ts, err := time.Parse(time.RFC3339Nano, text); err == nil {
				t.Time = ts.UnixNano()
			}
		}
		return nil
	}

	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		t.Time = n
	}

	return nil
}

func (t *Tick) Spread() float64 { return t.Ask - t.Bid }

func (t *Tick) Mid() float64 { return (t.Ask + t.Bid) / 2 }

// A buy trades at Ask, a sell at Bid.
func (t *Tick) OpenPrice(buy bool) float64 {
	if buy {
		return t.Ask
	}
	return t.Bid
}

func (t *Tick) ClosePrice(buyPosition bool) float64 {
	if buyPosition {
		return t.Bid
	}
	return t.Ask
}

// A missing side is not a price.
func (t *Tick) Ok() bool { return t.Bid > 0 && t.Ask > 0 }
