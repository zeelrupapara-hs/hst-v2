package model

import "time"

// Tick is a normalized quote tick published by hst-quote.
type Tick struct {
	DatafeedID int64     `json:"datafeed_id"`
	SymbolID   int64     `json:"symbol_id"`
	Symbol     string    `json:"symbol"`
	Source     string    `json:"source"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	High       float64   `json:"high,omitempty"`
	Low        float64   `json:"low,omitempty"`
	Open       float64   `json:"open,omitempty"`
	Close      float64   `json:"close,omitempty"`
	Volume     float64   `json:"volume,omitempty"`
	Time       time.Time `json:"time"`
}
