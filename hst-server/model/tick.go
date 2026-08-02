package model

import (
	"encoding/json"
	"strconv"
	"time"
)

// Tick is one quote, as hst-quote publishes it on hstquote.tick.<symbol>.
type Tick struct {
	Symbol     string  `json:"symbol"`
	Digits     int32   `json:"digits"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	Last       float64 `json:"last"`
	High       float64 `json:"high,omitempty"`
	Low        float64 `json:"low,omitempty"`
	Open       float64 `json:"open,omitempty"`
	Close      float64 `json:"close,omitempty"`
	Volume     int64   `json:"volume"`
	VolumeReal float64 `json:"volume_real"`
	// Time is when the quote was produced, epoch nanoseconds
	Time int64 `json:"time"`
}

// SubjectTickAll matches every symbol's quotes.
const SubjectTickAll = "hstquote.tick.*"

// UnmarshalJSON accepts either timestamp form, because refusing one drops every quote silently.
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

	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(w.Time, &text); err == nil {
			if ts, err := time.Parse(time.RFC3339Nano, text); err == nil {
				t.Time = ts.UnixNano()
			}
		}
		return nil
	}

	// read as an integer so the low digits survive
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		t.Time = n
	}

	return nil
}
