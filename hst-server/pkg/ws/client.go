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
type Client struct {
	// Id is unique per connection, so two sockets on one session are still distinguishable in the logs.
	Id string
	// SessionId is the session this socket authenticated with.
	SessionId string
	// Login owns the session.
	Login int64
	// Scope is the password slot the session authenticated with.
	Scope int32
	// Ip is the address resolved by the header reader.
	Ip string
	// IsManager is false for a trading account.
	IsManager bool
	// ConnectedAt is when the socket was accepted.
	ConnectedAt time.Time

	conn *websocket.Conn
	hub  *Hub
	log  *logger.Logger

	// egress is buffered: a publisher must never block on a slow client.
	egress chan *model.Event
	// closing is closed once, by close(), and is what every goroutine of this client selects on to stop.
	closing chan struct{}
	once    sync.Once

	// dropped counts events discarded because egress was full.
	dropped atomic.Int64

	// Groups is the group access these subscriptions were built from
	access sync.Mutex
	// subs are unsubscribed on close, so a disconnect leaks no consumer
	rights model.ManagerRights
	groups []string
	// subs are the nats subscriptions opened for this socket, unsubscribed on close so a disconnect does not leak a consumer.
	subs []Unsubscriber
}

// Access is the manager configuration these subscriptions were built from.
func (c *Client) Access() (model.ManagerRights, []string) {
	c.access.Lock()
	defer c.access.Unlock()

	groups := make([]string, len(c.groups))
	copy(groups, c.groups)
	return c.rights, groups
}

// Rights is the manager rights this socket holds.
func (c *Client) Rights() model.ManagerRights {
	rights, _ := c.Access()
	return rights
}

// Groups is the group access this socket holds.
func (c *Client) Groups() []string {
	_, groups := c.Access()
	return groups
}

// ReplaceSubs swaps the subscriptions for a new set and closes the old ones.
func (c *Client) ReplaceSubs(rights model.ManagerRights, groups []string, subs []Unsubscriber) {
	c.access.Lock()
	old := c.subs
	c.subs = subs
	c.rights = rights
	c.groups = groups
	c.access.Unlock()

	for _, s := range old {
		if err := s.Unsubscribe(); err != nil {
			c.log.Log(logger.TypeNet, logger.CodeWarn, "websocket unsubscribe failed",
				"session_id", c.SessionId, "error", err.Error())
		}
	}
}

// Unsubscriber is the part of a nats subscription this package needs.
type Unsubscriber interface {
	Unsubscribe() error
}

// Send queues an event.
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
		// one drop is a hiccup, a stream of them is a client that will never catch up; keeping it attached only wastes memory
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

// close tears the connection down exactly once.
func (c *Client) close() {
	c.once.Do(func() {
		close(c.closing)

		c.access.Lock()
		subs := c.subs
		c.subs = nil
		c.access.Unlock()

		for _, s := range subs {
			if err := s.Unsubscribe(); err != nil {
				c.log.Log(logger.TypeNet, logger.CodeWarn, "websocket unsubscribe failed",
					"session_id", c.SessionId, "error", err.Error())
			}
		}

		_ = c.conn.Close()
	})
}

// Wait blocks until the connection is torn down.
func (c *Client) Wait() { <-c.closing }

// writePump owns the connection's write side and the keepalive.
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

// readPump reads client frames.
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
		if err := json.Unmarshal(data, &in); err == nil && in.Type == model.EventPing {
			c.Send(&model.Event{Type: model.EventPong, At: time.Now().UnixNano()})
			continue
		}

		c.hub.Dispatch(c, data)
	}
}

// writeEvent puts one event on the wire in the frame its format asks for.
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
func encodeJSON(e *model.Event) []byte {
	buf := make([]byte, 0, len(e.Payload)+len(e.Type)+48)

	buf = append(buf, `{"type":`...)
	buf = appendQuoted(buf, e.Type)

	if e.Group != "" {
		buf = append(buf, `,"group":`...)
		buf = appendQuoted(buf, e.Group)
	}

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

// appendQuoted writes a json string.
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
