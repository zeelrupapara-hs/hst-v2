package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

// WaitTimeout is how long any wait on an event lasts before the test fails.
const WaitTimeout = 10 * time.Second

// Event is one frame as a client or a nats subscriber sees it.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Tick pushes a quote and, in dde mode, waits until hst-quote republished it on hstquote.tick.
func (e *Env) Tick(t *testing.T, symbol string, bid, ask float64) {
	t.Helper()
	if e.FeedMode == "nats" {
		raw, _ := json.Marshal(M{"symbol": symbol, "bid": bid, "ask": ask, "digits": 5, "time": time.Now().UTC()})
		if err := e.NC.Publish("hstquote.tick."+symbol, raw); err != nil {
			t.Fatalf("publish tick: %v", err)
		}
		_ = e.NC.Flush()
		return
	}
	// hst-quote drops a repeat of the same price inside one minute, so a price it already holds is done
	var last struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	}
	if raw, err := e.Redis.Get(context.Background(), "hstquote:last:"+symbol).Bytes(); err == nil {
		if json.Unmarshal(raw, &last) == nil && last.Bid == bid && last.Ask == ask {
			return
		}
	}
	sub, err := e.NC.SubscribeSync("hstquote.tick." + symbol)
	if err != nil {
		t.Fatalf("subscribe tick: %v", err)
	}
	defer sub.Unsubscribe()
	e.Feed.Step(symbol, bid, ask)
	deadline := time.Now().Add(WaitTimeout)
	for time.Now().Before(deadline) {
		msg, err := sub.NextMsg(time.Until(deadline))
		if err != nil {
			break
		}
		var tick struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		}
		if json.Unmarshal(msg.Data, &tick) == nil && tick.Bid == bid && tick.Ask == ask {
			return
		}
	}
	t.Fatalf("tick %s %.5f/%.5f never reached hstquote.tick (feed connected: %v)", symbol, bid, ask, e.Feed.Connected())
}

// Watcher buffers every event one account hears on nats from the moment it was made.
type Watcher struct {
	sub *nats.Subscription
	ch  chan Event
	// Seen is every event that was read off the subscription so far, for the failure message.
	Seen []string
}

// Watch subscribes to websocket.accounts.<login>.> before the action under test.
func (e *Env) Watch(t *testing.T, login int64) *Watcher {
	t.Helper()
	w := &Watcher{ch: make(chan Event, 1024)}
	sub, err := e.NC.Subscribe(fmt.Sprintf("websocket.accounts.%d.>", login), func(m *nats.Msg) {
		select {
		case w.ch <- Event{Type: m.Header.Get("X-Event"), Payload: m.Data}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("watch %d: %v", login, err)
	}
	w.sub = sub
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	return w
}

// Wait returns the first event of that type the predicate accepts, or fails the test.
func (w *Watcher) Wait(t *testing.T, kind string, pred func(payload []byte) bool) json.RawMessage {
	t.Helper()
	timer := time.NewTimer(WaitTimeout)
	defer timer.Stop()
	for {
		select {
		case ev := <-w.ch:
			w.Seen = append(w.Seen, ev.Type+" "+string(ev.Payload))
			if ev.Type == kind && (pred == nil || pred(ev.Payload)) {
				return ev.Payload
			}
		case <-timer.C:
			t.Fatalf("no %s event within %s; seen:\n%s", kind, WaitTimeout, strings.Join(w.Seen, "\n"))
		}
	}
}

// Summary is the account_summary line: summary,login,balance,credit,equity,margin,free,level%,profit.
type Summary struct {
	Balance, Credit, Equity, Margin, MarginFree, MarginLevel, Profit float64
}

// ParseSummary reads the text line the engine publishes on every money change.
func ParseSummary(raw []byte) (Summary, bool) {
	parts := strings.Split(string(raw), ",")
	if len(parts) < 9 || parts[0] != "summary" {
		return Summary{}, false
	}
	f := func(s string) float64 {
		v, _ := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		return v
	}
	return Summary{f(parts[2]), f(parts[3]), f(parts[4]), f(parts[5]), f(parts[6]), f(parts[7]), f(parts[8])}, true
}

// WaitSummary waits for an account_summary line the predicate accepts.
func (w *Watcher) WaitSummary(t *testing.T, pred func(Summary) bool) Summary {
	t.Helper()
	var out Summary
	w.Wait(t, "account_summary", func(raw []byte) bool {
		s, ok := ParseSummary(raw)
		if ok && pred(s) {
			out = s
			return true
		}
		return false
	})
	return out
}

// Field reads one json field of a payload as float64 (numbers and numeric ids).
func Field(t *testing.T, raw []byte, name string) float64 {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("payload is not an object: %s", raw)
	}
	v, ok := m[name].(float64)
	if !ok {
		t.Fatalf("payload has no number %q: %s", name, raw)
	}
	return v
}

// WS is one trader websocket connection to /ws.
type WS struct {
	conn *websocket.Conn
	seen []string
}

// WS opens the socket with the bearer subprotocol and waits for the welcome frame.
func (tr *Trader) WS(t *testing.T) *WS {
	t.Helper()
	url := "ws" + strings.TrimPrefix(tr.Client.Base, "http") + "/ws"
	d := websocket.Dialer{Subprotocols: []string{"bearer", tr.Client.Token.AccessToken}, HandshakeTimeout: 5 * time.Second}
	conn, resp, err := d.Dial(url, http.Header{})
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		t.Fatalf("ws dial: %v (status %d)", err, code)
	}
	w := &WS{conn: conn}
	t.Cleanup(func() { _ = conn.Close() })
	w.Wait(t, "welcome", nil)
	return w
}

// Send writes one {type,payload} frame.
func (w *WS) Send(t *testing.T, kind string, payload any) {
	t.Helper()
	raw, _ := json.Marshal(payload)
	if err := w.conn.WriteJSON(Event{Type: kind, Payload: raw}); err != nil {
		t.Fatalf("ws send %s: %v", kind, err)
	}
}

// Wait reads frames until one of that type passes the predicate; bad_request and friends fail at once.
func (w *WS) Wait(t *testing.T, kind string, pred func(payload []byte) bool) json.RawMessage {
	t.Helper()
	deadline := time.Now().Add(WaitTimeout)
	for time.Now().Before(deadline) {
		_ = w.conn.SetReadDeadline(deadline)
		_, raw, err := w.conn.ReadMessage()
		if err != nil {
			break
		}
		var ev Event
		if json.Unmarshal(raw, &ev) != nil {
			continue
		}
		w.seen = append(w.seen, string(raw))
		if ev.Type == kind && (pred == nil || pred(ev.Payload)) {
			return ev.Payload
		}
		switch ev.Type {
		case "bad_request", "forbidden", "not_found", "unauthorized", "internal_server_error":
			t.Fatalf("ws refused while waiting for %s: %s", kind, raw)
		}
	}
	t.Fatalf("no %s frame on the socket within %s; seen:\n%s", kind, WaitTimeout, strings.Join(w.seen, "\n"))
	return nil
}
