package v1

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

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
}

var feed = &marketFeed{clients: make(map[string]*marketWatcher, 128)}

// StartMyMarketFeed puts the caller on the price stream.
func (s *HttpServer) StartMyMarketFeed(c *ws.Ctx) error {
	granted, err := s.grantedSymbols(c)
	if err != nil {
		return err
	}

	feed.mu.Lock()
	feed.clients[c.Client.SessionId] = &marketWatcher{client: c.Client, symbols: granted}
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

	line := TickLine(&t)

	feed.mu.RLock()
	defer feed.mu.RUnlock()

	for _, w := range feed.clients {
		// an instrument the account's group was never granted is not one it may be quoted
		if w.symbols != nil && !w.symbols[t.Symbol] {
			continue
		}

		w.client.Send(&model.Event{Type: model.EventMarketFeed, Payload: []byte(line), Format: model.FormatBinary})
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
