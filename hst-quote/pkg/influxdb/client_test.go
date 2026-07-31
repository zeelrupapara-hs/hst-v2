package influxdb

import (
	"testing"
	"time"

	"hstquote/model"
)

func TestWriteTickNilClientNoPanic(t *testing.T) {
	var c *Client
	c.WriteTick(model.Tick{SymbolID: 42, Symbol: "EURUSD", Bid: 1.1, Ask: 1.2, Time: time.Now().UTC()})
}

func TestWriteTickSkipsMissingSymbolID(t *testing.T) {
	c := &Client{writer: nil}
	c.WriteTick(model.Tick{Symbol: "EURUSD", Bid: 1.1, Ask: 1.2, Time: time.Now().UTC()})
}

func TestPointFromTickUsesSymbolIDMeasurement(t *testing.T) {
	ts := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	point := pointFromTick(model.Tick{
		SymbolID: 42,
		Symbol:   "EURUSD",
		Bid:      1.1,
		Ask:      1.2,
		High:     1.25,
		Low:      1.05,
		Open:     1.11,
		Close:    1.15,
		Volume:   100,
		Time:     ts,
	})

	if point.Name() != "42" {
		t.Fatalf("measurement: %q", point.Name())
	}
	if got := point.Time(); !got.Equal(ts) {
		t.Fatalf("time: %v", got)
	}

	fields := map[string]any{
		"bid": 1.1, "ask": 1.2, "high": 1.25, "low": 1.05,
		"open": 1.11, "close": 1.15, "volume": float64(100),
	}
	gotFields := make(map[string]any, len(fields))
	for _, f := range point.FieldList() {
		gotFields[f.Key] = f.Value
	}
	for name, want := range fields {
		got, ok := gotFields[name]
		if !ok {
			t.Fatalf("missing field %q", name)
		}
		if got != want {
			t.Fatalf("field %q: got %v want %v", name, got, want)
		}
	}
	if len(point.TagList()) != 0 {
		t.Fatalf("expected no tags, got %v", point.TagList())
	}
}
