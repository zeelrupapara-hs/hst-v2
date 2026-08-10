package v1

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"
	"hstserver/pkg/ws"

	natscore "github.com/nats-io/nats.go"
)

// One process-wide subscription fans ticks out from memory, and only to terminals that asked.

// marketFeed is who is watching prices, and which instruments each may see.
type marketFeed struct {
	mu      sync.RWMutex
	clients map[string]*marketWatcher
}

// marketWatcher is one terminal on the feed.
type marketWatcher struct {
	client *ws.Client
	// symbols is what this login's group grants it; nil means staff, scoped to no group list
	symbols map[string]bool
	// group prices the stream: its spread difference is added before the tick leaves
	group string
}

var feed = &marketFeed{clients: make(map[string]*marketWatcher, 128)}

// groupSpread is one group's price transform on one symbol: the difference is split around the
// mid and the balance shifts the split, exactly as the engine charges it.
type groupSpread struct {
	diff    int32
	balance int32
	digits  int32
}

// spreadBook is every group's spread transform, reloaded whenever symbols or overrides change.
type spreadBook struct {
	mu sync.RWMutex
	// by group, then symbol; only symbols with a non-zero transform are held
	groups map[string]map[string]groupSpread
}

var spreads = &spreadBook{groups: make(map[string]map[string]groupSpread, 16)}

func (b *spreadBook) forGroup(group, symbol string) (groupSpread, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	gs, ok := b.groups[group][symbol]
	return gs, ok
}

// loadSpreadBook resolves each group's spread pair the way the engine does: the first override
// row in config order wins, an absent field inherits the symbol's own value.
func (s *HttpServer) loadSpreadBook() {
	rows, err := s.DB.DB.Query(context.Background(),
		`SELECT g."group", sym.symbol, sym.digits,
		        COALESCE(o.spread_diff, sym.spread_diff),
		        COALESCE(o.spread_diff_balance, sym.spread_diff_balance)
		   FROM hst.groups g
		   JOIN hst.symbols sym ON TRUE
		   JOIN LATERAL (
		        SELECT gs.spread_diff, gs.spread_diff_balance
		          FROM hst.groups_symbols gs
		         WHERE gs.group_id = g.group_id
		           AND (gs.path = '*' OR sym.path = gs.path OR sym.symbol = gs.path
		                OR starts_with(sym.path, rtrim(gs.path, '*')))
		         ORDER BY gs.config_index
		         LIMIT 1) o ON TRUE`)
	if err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "spread book load failed", "error", err.Error())
		return
	}
	defer rows.Close()

	next := make(map[string]map[string]groupSpread, 16)
	for rows.Next() {
		var group, symbol string
		var digits, diff, balance int32
		if err := rows.Scan(&group, &symbol, &digits, &diff, &balance); err != nil {
			s.Log.Log(logger.TypeNet, logger.CodeWarn, "spread book scan failed", "error", err.Error())
			return
		}
		if diff == 0 && balance == 0 {
			continue
		}
		if next[group] == nil {
			next[group] = make(map[string]groupSpread, 8)
		}
		next[group][symbol] = groupSpread{diff: diff, balance: balance, digits: digits}
	}

	spreads.mu.Lock()
	spreads.groups = next
	spreads.mu.Unlock()
}

// applyGroupSpread is the engine's CalculateAccountSpread, applied to the outgoing stream so a
// terminal renders the exact price its orders fill at.
func applyGroupSpread(gs groupSpread, t model.Tick) model.Tick {
	pt := math.Pow(10, -float64(gs.digits))
	d := float64(gs.diff)
	b := float64(gs.balance)

	t.Bid = roundTo(t.Bid-(d/2-b)*pt, gs.digits)
	t.Ask = roundTo(t.Ask+(d/2+b)*pt, gs.digits)
	return t
}

func roundTo(v float64, digits int32) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}

// symLiveness knows when each symbol last ticked; the panel greys the ones that fell silent.
type symLiveness struct {
	mu   sync.Mutex
	last map[string]int64
	live map[string]bool
}

var symLive = &symLiveness{last: make(map[string]int64, 128), live: make(map[string]bool, 128)}

// symbolStaleFloor guards against a feed whose timeout is unset: 0 would grey everything.
const symbolStaleFloor = 30 * time.Second

const symbolLivenessSweep = 10 * time.Second

// Touch records that a symbol just ticked.
func (l *symLiveness) Touch(symbol string, timeNs int64) {
	l.mu.Lock()
	if timeNs > l.last[symbol] {
		l.last[symbol] = timeNs
	}
	l.mu.Unlock()
}

// StartSymbolLiveness seeds last-tick times from the quote cache and keeps the panel told.
func (s *HttpServer) StartSymbolLiveness() {
	s.seedSymbolLiveness()

	// work the map out once now: a panel connecting before the first sweep would otherwise be
	// handed nothing and grey every symbol, including the ones already ticking
	if _, err := s.refreshSymbolLiveness(context.Background()); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "first symbol liveness pass failed",
			"error", err.Error())
	}

	go s.sweepSymbolLiveness()
}

// Snapshot copies what is currently live, for a socket that has just connected.
func (l *symLiveness) Snapshot() map[string]bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make(map[string]bool, len(l.live))
	for symbol, live := range l.live {
		out[symbol] = live
	}

	return out
}

// SendSymbolLiveness tells one socket what is ticking, as soon as it connects. The sweep only
// speaks every ten seconds, which is ten seconds of a freshly loaded panel showing grey.
func (s *HttpServer) SendSymbolLiveness(c *ws.Client) {
	live := symLive.Snapshot()
	if len(live) == 0 {
		return
	}

	raw, err := json.Marshal(map[string]any{"live": live})
	if err != nil {
		return
	}

	c.Send(&model.Event{
		Type:    model.EventSymbolLivenessUpdated,
		Format:  model.FormatJSON,
		Payload: raw,
		At:      time.Now().UnixNano(),
	})
}

// refreshSymbolLiveness recomputes which symbols are still inside their feed's timeout.
func (s *HttpServer) refreshSymbolLiveness(ctx context.Context) (map[string]bool, error) {
	bounds, err := s.symbolStaleBounds(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().UnixNano()
	next := make(map[string]bool, len(bounds))

	symLive.mu.Lock()
	for symbol, bound := range bounds {
		next[symbol] = now-symLive.last[symbol] <= bound.Nanoseconds()
	}
	symLive.live = next
	symLive.mu.Unlock()

	return next, nil
}

// the quote service keeps every symbol's last tick in redis; a restart must not grey the world
func (s *HttpServer) seedSymbolLiveness() {
	ctx := context.Background()
	iter := s.Redis.Client.Scan(ctx, 0, "hstquote:last:*", 500).Iterator()

	seeded := 0
	for iter.Next(ctx) {
		raw, err := s.Redis.Client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		var t model.Tick
		if json.Unmarshal([]byte(raw), &t) != nil || t.Symbol == "" {
			continue
		}
		symLive.Touch(t.Symbol, t.Time)
		seeded++
	}

	s.Log.Log(logger.TypeNet, logger.CodeOK, "symbol liveness seeded", "symbols", seeded)
}

// symbolStaleBounds is each symbol with how long its serving feed waits for a quote.
func (s *HttpServer) symbolStaleBounds(ctx context.Context) (map[string]time.Duration, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT s.symbol,
		        COALESCE((SELECT MAX(d.timeout)
		                    FROM hst.datafeeds d
		                   WHERE d.enable = 1
		                     AND EXISTS (SELECT 1
		                                   FROM hst.datafeed_symbols ds
		                                  WHERE ds.datafeed_id = d.datafeed_id
		                                    AND ds.exclude = 0
		                                    AND (ds.path = '*'
		                                         OR ds.symbol = s.symbol
		                                         OR s.path = ds.path
		                                         OR starts_with(s.path, rtrim(ds.path, '*'))))), 0)
		   FROM hst.symbols s`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]time.Duration, 128)
	for rows.Next() {
		var symbol string
		var timeout int32
		if err := rows.Scan(&symbol, &timeout); err != nil {
			return nil, err
		}
		bound := time.Duration(timeout) * time.Second
		if bound < symbolStaleFloor {
			bound = symbolStaleFloor
		}
		out[symbol] = bound
	}

	return out, rows.Err()
}

// sweepSymbolLiveness recomputes who is live and pushes the whole map; a fresh terminal has the
// full picture within one sweep, so there is no snapshot call to serve.
func (s *HttpServer) sweepSymbolLiveness() {
	ticker := time.NewTicker(symbolLivenessSweep)
	defer ticker.Stop()

	for range ticker.C {
		next, err := s.refreshSymbolLiveness(context.Background())
		if err != nil {
			s.Log.Log(logger.TypeNet, logger.CodeWarn, "symbol liveness sweep failed", "error", err.Error())
			continue
		}

		s.NotifyWS(model.SubjectSymbol, model.EventSymbolLivenessUpdated, map[string]any{"live": next})
	}
}

// StartMyMarketFeed puts the caller on the price stream.
func (s *HttpServer) StartMyMarketFeed(c *ws.Ctx) error {
	granted, err := s.grantedSymbols(c)
	if err != nil {
		return err
	}

	group := ""
	if !c.Client.IsManager {
		group = s.groupOf(c.Client.Login)
	}

	feed.mu.Lock()
	feed.clients[c.Client.SessionId] = &marketWatcher{client: c.Client, symbols: granted, group: group}
	watching := len(feed.clients)
	feed.mu.Unlock()

	s.Log.Log(logger.TypeNet, logger.CodeOK, "market feed started",
		"login", c.Client.Login, "symbols", len(granted), "watching", watching)

	return c.SendEvent(&model.Event{Type: model.EventStartMarketFeed, Format: model.FormatJSON,
		Payload: []byte(`"market feed has started"`)})
}

// StopMyMarketFeed takes the caller off the price stream.
func (s *HttpServer) StopMyMarketFeed(c *ws.Ctx) error {
	dropMarketWatcher(c.Client.SessionId)

	return c.SendEvent(&model.Event{Type: model.EventStopMarketFeed, Format: model.FormatJSON,
		Payload: []byte(`"market feed has stopped"`)})
}

// dropMarketWatcher forgets a terminal, on request or when its socket closes.
func dropMarketWatcher(sessionId string) {
	feed.mu.Lock()
	delete(feed.clients, sessionId)
	feed.mu.Unlock()
}

// MarketFeedHandler takes one tick off the wire and hands it to everyone entitled to see it.
func (s *HttpServer) MarketFeedHandler(msg *natscore.Msg) {
	var t model.Tick
	if err := json.Unmarshal(msg.Data, &t); err != nil {
		return
	}

	s.AlertsOnTick(&t)

	symLive.Touch(t.Symbol, t.Time)

	line := TickLine(&t)
	// one adjusted line per group on this tick; most groups share the raw line
	byGroup := map[string]string{"": line}

	feed.mu.RLock()
	defer feed.mu.RUnlock()

	for _, w := range feed.clients {
		// an instrument the account's group was never granted is not one it may be quoted
		if w.symbols != nil && !w.symbols[t.Symbol] {
			continue
		}

		out, ok := byGroup[w.group]
		if !ok {
			if gs, has := spreads.forGroup(w.group, t.Symbol); has {
				adj := applyGroupSpread(gs, t)
				out = TickLine(&adj)
			} else {
				out = line
			}
			byGroup[w.group] = out
		}

		w.client.Send(&model.Event{Type: model.EventMarketFeed, Payload: []byte(out), Format: model.FormatBinary})
	}
}

// TickLine is one quote as the terminal reads it: symbol,bid,ask,last,volume,ts,open,high,low,close,change,change_percent
func TickLine(t *model.Tick) string {
	var b strings.Builder

	d := int(t.Digits)
	if d <= 0 {
		d = 5
	}

	b.WriteString(t.Symbol)

	for _, v := range []float64{t.Bid, t.Ask, t.Last} {
		b.WriteByte(',')
		b.WriteString(strconv.FormatFloat(v, 'f', d, 64))
	}

	b.WriteByte(',')
	b.WriteString(strconv.FormatInt(t.Volume, 10))
	b.WriteByte(',')
	b.WriteString(strconv.FormatInt(t.Time/1e9, 10))

	var change, changePct float64
	if t.Close != 0 {
		change = t.Last - t.Close
		changePct = change / t.Close * 100
	}

	for _, v := range []float64{t.Open, t.High, t.Low, t.Close, change, changePct} {
		b.WriteByte(',')
		b.WriteString(strconv.FormatFloat(v, 'f', d, 64))
	}

	return b.String()
}

// grantedSymbols is what this caller may be quoted, or nil for a member of staff.
func (s *HttpServer) grantedSymbols(c *ws.Ctx) (map[string]bool, error) {
	if c.Client.IsManager {
		return nil, nil
	}

	rows, err := s.DB.DB.Query(context.Background(),
		`SELECT s.symbol
		   FROM hst.symbols s
		  WHERE EXISTS (
		        SELECT 1
		          FROM hst.groups_symbols gs
		          JOIN hst.groups g ON g.group_id = gs.group_id
		          JOIN hst.users u ON u."group" = g."group"
		         WHERE u.login = $1
		           AND (gs.path = '*' OR s.path = gs.path OR starts_with(s.path, rtrim(gs.path, '*')))
		  )`, c.Client.Login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool, 128)

	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		out[symbol] = true
	}

	return out, rows.Err()
}
