// Package influxdb reads back the ticks hst-quote wrote, as candles.
package influxdb

import (
	"context"
	"fmt"
	"time"

	"hstserver/config"
	"hstserver/pkg/logger"

	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// Reader answers chart history. It never writes.
type Reader struct {
	client influx.Client
	query  api.QueryAPI
	// ticks holds the raw stream for a short window; candles holds the minute bars for good.
	ticks     string
	candles   string
	tickReach time.Duration
	Log       *logger.Logger
}

// Candle is one bar.
type Candle struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// NewReader connects, or reports that history is switched off.
func NewReader(cfg *config.Config, log *logger.Logger) (*Reader, error) {
	if !cfg.Influx.Enabled {
		log.Log(logger.TypeSys, logger.CodeWarn, "chart history is disabled, no tick store configured")
		return nil, nil
	}

	if cfg.Influx.Url == "" || cfg.Influx.Token == "" {
		return nil, fmt.Errorf("influx url and token are required when history is enabled")
	}

	client := influx.NewClient(cfg.Influx.Url, cfg.Influx.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("influx ping: %w", err)
	}

	log.Log(logger.TypeSys, logger.CodeOK, "chart history connected",
		"org", cfg.Influx.Org, "bucket", cfg.Influx.Bucket)

	return &Reader{
		client:    client,
		query:     client.QueryAPI(cfg.Influx.Org),
		ticks:     cfg.Influx.Bucket,
		candles:   cfg.Influx.CandleBucket,
		tickReach: cfg.Influx.TickRetention,
		Log:       log,
	}, nil
}

func (r *Reader) Close() {
	if r != nil && r.client != nil {
		r.client.Close()
	}
}

// Candles builds bars of one width for one instrument.
//
// There are two places the answer can come from. The raw ticks are the more exact of the two and
// carry the bar that is still being formed, but they are only kept for a short window. Behind
// that window the minute bars the rollup wrote are all that is left, and longer bars are built
// out of those. Which one is used depends on how far back the request reaches.
//
// The bid is the chart price, as it is in the terminal: it is the side a long position is valued
// and closed at, so it is the line a trader is actually watching.
func (r *Reader) Candles(ctx context.Context, symbolId int64, every string,
	from, to time.Time) ([]Candle, error) {
	if r == nil {
		return nil, ErrHistoryDisabled
	}

	if r.withinTicks(from) {
		return r.fromTicks(ctx, symbolId, every, from, to)
	}

	return r.fromCandles(ctx, symbolId, every, from, to)
}

// withinTicks reports whether the raw stream still reaches back this far. An hour is left at the
// edge so a request that straddles the boundary is not served half-empty.
func (r *Reader) withinTicks(from time.Time) bool {
	if r.tickReach <= 0 {
		return false
	}

	return from.After(time.Now().Add(-r.tickReach + time.Hour))
}

// fromTicks cuts bars straight out of the tick stream.
func (r *Reader) fromTicks(ctx context.Context, symbolId int64, every string,
	from, to time.Time) ([]Candle, error) {
	flux := fmt.Sprintf(`
bid = from(bucket: %q)
  |> range(start: %d, stop: %d)
  |> filter(fn: (r) => r._measurement == "%d" and r._field == "bid")

vol = from(bucket: %q)
  |> range(start: %d, stop: %d)
  |> filter(fn: (r) => r._measurement == "%d" and r._field == "volume")

o = bid |> aggregateWindow(every: %s, fn: first, timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "o")
h = bid |> aggregateWindow(every: %s, fn: max,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "h")
l = bid |> aggregateWindow(every: %s, fn: min,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "l")
c = bid |> aggregateWindow(every: %s, fn: last,  timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "c")
v = vol |> aggregateWindow(every: %s, fn: sum,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "v")

union(tables: [o, h, l, c, v])
  |> keep(columns: ["_time", "_value", "b"])
  |> pivot(rowKey: ["_time"], columnKey: ["b"], valueColumn: "_value")
  |> sort(columns: ["_time"])
`,
		r.ticks, from.Unix(), to.Unix(), symbolId,
		r.ticks, from.Unix(), to.Unix(), symbolId,
		every, every, every, every, every)

	return r.collect(ctx, flux)
}

// fromCandles builds longer bars out of the minute bars, which is all that survives the raw
// window. A minute request passes straight through.
func (r *Reader) fromCandles(ctx context.Context, symbolId int64, every string,
	from, to time.Time) ([]Candle, error) {
	flux := fmt.Sprintf(`
base = from(bucket: %q)
  |> range(start: %d, stop: %d)
  |> filter(fn: (r) => r._measurement == "%d")

o = base |> filter(fn: (r) => r._field == "open")   |> aggregateWindow(every: %s, fn: first, timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "o")
h = base |> filter(fn: (r) => r._field == "high")   |> aggregateWindow(every: %s, fn: max,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "h")
l = base |> filter(fn: (r) => r._field == "low")    |> aggregateWindow(every: %s, fn: min,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "l")
c = base |> filter(fn: (r) => r._field == "close")  |> aggregateWindow(every: %s, fn: last,  timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "c")
v = base |> filter(fn: (r) => r._field == "volume") |> aggregateWindow(every: %s, fn: sum,   timeSrc: "_start", createEmpty: false) |> set(key: "b", value: "v")

union(tables: [o, h, l, c, v])
  |> keep(columns: ["_time", "_value", "b"])
  |> pivot(rowKey: ["_time"], columnKey: ["b"], valueColumn: "_value")
  |> sort(columns: ["_time"])
`,
		r.candles, from.Unix(), to.Unix(), symbolId,
		every, every, every, every, every)

	return r.collect(ctx, flux)
}

// collect turns a pivoted result into bars.
func (r *Reader) collect(ctx context.Context, flux string) ([]Candle, error) {
	rows, err := r.query.Query(ctx, flux)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Candle, 0, 512)

	for rows.Next() {
		rec := rows.Record()

		bar := Candle{
			Time:   rec.Time().Unix(),
			Open:   number(rec.ValueByKey("o")),
			High:   number(rec.ValueByKey("h")),
			Low:    number(rec.ValueByKey("l")),
			Close:  number(rec.ValueByKey("c")),
			Volume: number(rec.ValueByKey("v")),
		}

		// a window with no price in it is not a bar, whatever else landed in it
		if bar.Open == 0 && bar.Close == 0 {
			continue
		}

		out = append(out, bar)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return out, nil
}

func number(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	}

	return 0
}

// ErrHistoryDisabled is returned when no tick store is configured.
var ErrHistoryDisabled = fmt.Errorf("chart history is not configured")

// FirstTick is when this instrument's history begins, so a chart knows when to stop asking for
// more of it.
func (r *Reader) FirstTick(ctx context.Context, symbolId int64) (time.Time, error) {
	if r == nil {
		return time.Time{}, ErrHistoryDisabled
	}

	flux := fmt.Sprintf(`
from(bucket: %q)
  |> range(start: 0)
  |> filter(fn: (r) => r._measurement == "%d" and r._field == "open")
  |> first()
  |> keep(columns: ["_time"])
`, r.candles, symbolId)

	rows, err := r.query.Query(ctx, flux)
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Record().Time(), nil
	}

	return time.Time{}, rows.Err()
}
