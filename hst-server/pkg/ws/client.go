package ws

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
)

// Client is one websocket connection.
//
// Every write goes through egress and the single writer goroutine. The
// websocket library forbids concurrent writes, so serialising them is not a
// convenience here, it is the only thing keeping the connection from being
// corrupted by two publishers at once.
type Client struct {
	// Id is unique per connection, so two sockets on one session are still
	// distinguishable in the logs.
	Id string
	// SessionId is the session this socket authenticated with.
	SessionId string
	// Login owns the session.
	Login int64
	// Ip is the address resolved by the header reader.
	Ip string
	// Rights are the manager rights the session held at connect time.
	Rights model.ManagerRights
	// IsManager is false for a trading account.
	IsManager bool
	// ConnectedAt is when the socket was accepted.
	ConnectedAt time.Time

	conn *websocket.Conn
	hub  *Hub
	log  *logger.Logger

	// egress is buffered: a publisher must never block on a slow client.
	egress chan *model.Event
	// closing is closed once, by close(), and is what every goroutine of this
	// client selects on to stop.
	closing chan struct{}
	once    sync.Once

	// dropped counts events discarded because egress was full. A client that
	// keeps dropping is disconnected rather than left silently lossy.
	dropped atomic.Int64
	// subs are the nats subscriptions opened for this socket, unsubscribed on
	// close so a disconnect does not leak a consumer.
	subs []Unsubscriber
}

// Unsubscriber is the part of a nats subscription this package needs. Keeping
// it an interface means pkg/ws does not import nats at all.
type Unsubscriber interface {
	Unsubscribe() error
}

// Send queues an event. It never blocks: a websocket client that cannot keep
// up must not be able to stall the nats callback that is feeding it, because
// that callback is shared with every other subscriber on the same connection.
//
// It reports false when the event was dropped.
func (c *Client) Send(e *model.Event) bool {
	select {
	case <-c.closing:
		return false
	default:
	}

	select {
	case c.egress <- e:
		return true
	default:
		n := c.dropped.Add(1)
		// one drop is a hiccup, a stream of them is a client that will never
		// catch up; keeping it attached only wastes memory
		if n >= maxDrops {
			c.log.Log(logger.TypeNet, logger.CodeWarn, "closing a websocket that cannot keep up",
				"session_id", c.SessionId, "login", c.Login, "dropped", n)
			c.hub.Remove(c.SessionId, c.Id)
			return false
		}
		return false
	}
}

// Dropped is how many events this socket has lost, worth exporting as a metric.
func (c *Client) Dropped() int64 { return c.dropped.Load() }

// AddSub records a subscription to be closed with the connection.
func (c *Client) AddSub(s Unsubscriber) {
	if s != nil {
		c.subs = append(c.subs, s)
	}
}

// close tears the connection down exactly once.
func (c *Client) close() {
	c.once.Do(func() {
		close(c.closing)

		for _, s := range c.subs {
			if err := s.Unsubscribe(); err != nil {
				c.log.Log(logger.TypeNet, logger.CodeWarn, "websocket unsubscribe failed",
					"session_id", c.SessionId, "error", err.Error())
			}
		}

		_ = c.conn.Close()
	})
}

// Wait blocks until the connection is torn down. The fiber handler has to sit
// here: the socket is closed the moment the handler returns.
func (c *Client) Wait() { <-c.closing }

// writePump owns the connection's write side and the keepalive. Nothing else
// may write to the socket.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		// a write failure means this connection is finished
		c.hub.Remove(c.SessionId, c.Id)
	}()

	for {
		select {
		case <-c.closing:
			return

		case e := <-c.egress:
			if err := c.writeEvent(e); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump reads client frames. Nothing inbound is routed anywhere yet, but
// the read has to run regardless: it is what drives the pong handler, and
// without it a half open connection is never noticed.
func (c *Client) readPump() {
	defer c.hub.Remove(c.SessionId, c.Id)

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				c.log.Log(logger.TypeNet, logger.CodeWarn, "websocket read failed",
					"session_id", c.SessionId, "error", err.Error())
			}
			return
		}

		// a client with no native ping support can send a ping event instead
		var in model.Event
		if err := json.Unmarshal(data, &in); err != nil {
			continue
		}
		if in.Type == model.EventPing {
			c.Send(&model.Event{Type: model.EventPong, At: time.Now().UnixNano()})
		}
	}
}

// writeEvent puts one event on the wire in the frame its format asks for.
//
// Binary and text are forwarded verbatim: no parse, no envelope, no copy. That
// is the whole point of letting a publisher choose the format, and it is what
// makes a packed tick stream affordable.
func (c *Client) writeEvent(e *model.Event) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

	switch e.Format {
	case model.FormatBinary:
		return c.conn.WriteMessage(websocket.BinaryMessage, e.Payload)
	case model.FormatText:
		return c.conn.WriteMessage(websocket.TextMessage, e.Payload)
	default:
		return c.conn.WriteMessage(websocket.TextMessage, encodeJSON(e))
	}
}

// encodeJSON builds {"type":..,"at":..,"payload":<raw>} by hand.
//
// The payload is already encoded, so running it back through a marshaller
// would mean parsing json only to print the same json again. Appending bytes
// skips the reflection entirely, and at a few thousand events a second that is
// the difference worth having.
func encodeJSON(e *model.Event) []byte {
	buf := make([]byte, 0, len(e.Payload)+len(e.Type)+48)

	buf = append(buf, `{"type":`...)
	buf = appendQuoted(buf, e.Type)

	if e.At != 0 {
		buf = append(buf, `,"at":`...)
		buf = strconv.AppendInt(buf, e.At, 10)
	}

	if len(e.Payload) > 0 {
		buf = append(buf, `,"payload":`...)
		buf = append(buf, e.Payload...)
	}

	return append(buf, '}')
}

// appendQuoted writes a json string. Event names come from nats subjects, so
// they are plain ascii in every real case; the marshaller is only there to
// keep a strange one from producing broken json.
func appendQuoted(buf []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if b := s[i]; b < 0x20 || b == '"' || b == '\\' || b > 0x7e {
			raw, err := json.Marshal(s)
			if err != nil {
				return append(buf, `""`...)
			}
			return append(buf, raw...)
		}
	}

	buf = append(buf, '"')
	buf = append(buf, s...)
	return append(buf, '"')
}
