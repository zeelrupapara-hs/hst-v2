package v1

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	errs "hstserver/pkg/errors"
	"hstserver/pkg/influxdb"

	"github.com/jackc/pgx/v5"
)

// Chart history.
//
// hst-quote writes every tick it receives into the tick store; this reads them back as bars. The
// answer is shaped the way a charting library expects it — parallel arrays rather than a list of
// objects — because that is what TradingView's datafeed contract asks for and it is a good deal
// smaller on the wire for a few thousand bars.

// Bars is a chart history answer.
type Bars struct {
	Status string    `json:"s"`
	Time   []int64   `json:"t,omitempty"`
	Open   []float64 `json:"o,omitempty"`
	High   []float64 `json:"h,omitempty"`
	Low    []float64 `json:"l,omitempty"`
	Close  []float64 `json:"c,omitempty"`
	Volume []float64 `json:"v,omitempty"`
	// NextTime tells the chart where the data actually starts, so it stops asking further back.
	NextTime int64 `json:"nextTime,omitempty"`
}

// HistoryQuery is one request for bars.
type HistoryQuery struct {
	Symbol     string
	Resolution string
	From       time.Time
	To         time.Time
	Countback  int
}

// maxBars is as many as one answer will carry, so a careless range cannot pull the whole store.
const maxBars = 10000

// resolutions maps what a chart asks for onto a window the tick store understands. The keys are
// the charting library's own names: minutes as a bare number, then D, W and M.
var resolutions = map[string]string{
	"1": "1m", "3": "3m", "5": "5m", "15": "15m", "30": "30m",
	"45": "45m", "60": "1h", "120": "2h", "180": "3h", "240": "4h",
	"1D": "1d", "D": "1d", "1W": "1w", "W": "1w", "1M": "1mo", "M": "1mo",
}

// readHistoryQuery pulls a chart's request off the query string.
func readHistoryQuery(symbol, resolution, from, to, countback string) (*HistoryQuery, error) {
	q := &HistoryQuery{
		Symbol:     strings.TrimSpace(symbol),
		Resolution: strings.TrimSpace(resolution),
	}

	if q.Symbol == "" {
		return nil, errs.ErrRequiredParams
	}

	if q.Resolution == "" {
		q.Resolution = "1"
	}
	if _, ok := resolutions[q.Resolution]; !ok {
		return nil, errs.ErrInvalidResolution
	}

	sec, err := strconv.ParseInt(from, 10, 64)
	if err != nil || sec <= 0 {
		return nil, errs.ErrRequiredParams
	}
	q.From = time.Unix(sec, 0).UTC()

	sec, err = strconv.ParseInt(to, 10, 64)
	if err != nil || sec <= 0 {
		return nil, errs.ErrRequiredParams
	}
	q.To = time.Unix(sec, 0).UTC()

	if !q.To.After(q.From) {
		return nil, errs.ErrInvalidRange
	}

	if countback != "" {
		if n, err := strconv.Atoi(countback); err == nil && n > 0 {
			q.Countback = n
		}
	}

	return q, nil
}

// history answers one chart request against a symbol the caller has already been cleared for.
func (s *HttpServer) history(ctx context.Context, symbolId int64, q *HistoryQuery) (*Bars, error) {
	candles, err := s.History.Candles(ctx, symbolId, resolutions[q.Resolution], q.From, q.To)
	if err != nil {
		return nil, err
	}

	if len(candles) == 0 {
		out := &Bars{Status: "no_data"}

		// point the chart at the first tick we hold, so it stops walking backwards forever
		if first, err := s.History.FirstTick(ctx, symbolId); err == nil && !first.IsZero() {
			if first.After(q.To) || first.After(q.From) {
				out.NextTime = first.Unix()
			}
		}

		return out, nil
	}

	// a chart asking for a bar count wants the newest ones
	if q.Countback > 0 && len(candles) > q.Countback {
		candles = candles[len(candles)-q.Countback:]
	}
	if len(candles) > maxBars {
		candles = candles[len(candles)-maxBars:]
	}

	out := &Bars{
		Status: "ok",
		Time:   make([]int64, 0, len(candles)),
		Open:   make([]float64, 0, len(candles)),
		High:   make([]float64, 0, len(candles)),
		Low:    make([]float64, 0, len(candles)),
		Close:  make([]float64, 0, len(candles)),
		Volume: make([]float64, 0, len(candles)),
	}

	for _, b := range candles {
		out.Time = append(out.Time, b.Time)
		out.Open = append(out.Open, b.Open)
		out.High = append(out.High, b.High)
		out.Low = append(out.Low, b.Low)
		out.Close = append(out.Close, b.Close)
		out.Volume = append(out.Volume, b.Volume)
	}

	return out, nil
}

// symbolForTrader resolves a symbol name for an account, and refuses one its group was never
// granted. The same rule the symbol list is built from decides it.
func (s *HttpServer) symbolForTrader(ctx context.Context, login int64, symbol string) (int64, error) {
	var id int64

	err := s.DB.DB.QueryRow(ctx,
		`SELECT s.symbol_id
		   FROM hst.symbols s
		  WHERE s.symbol = $2
		    AND EXISTS (
		        SELECT 1
		          FROM hst.groups_symbols gs
		          JOIN hst.groups g ON g.group_id = gs.group_id
		          JOIN hst.users u ON u."group" = g."group"
		         WHERE u.login = $1
		           AND (gs.path = '*' OR s.path = gs.path OR starts_with(s.path, rtrim(gs.path, '*')))
		    )`, login, symbol).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errs.ErrNotFound
	}

	return id, err
}

// symbolByName resolves a symbol for a member of staff, who is not scoped to a group's list.
func (s *HttpServer) symbolByName(ctx context.Context, symbol string) (int64, error) {
	var id int64

	err := s.DB.DB.QueryRow(ctx,
		`SELECT symbol_id FROM hst.symbols WHERE symbol = $1`, symbol).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errs.ErrNotFound
	}

	return id, err
}

// historyUnavailable reports whether the failure was simply that no tick store is configured.
func historyUnavailable(err error) bool {
	return errors.Is(err, influxdb.ErrHistoryDisabled)
}
