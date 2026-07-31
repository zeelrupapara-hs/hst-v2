package ws

import (
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
			if err := c.writeJSON(e); err != nil {
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

func (c *Client) writeJSON(e *model.Event) error {
	raw, err := json.Marshal(e)
	if err != nil {
		// a payload this server cannot encode is the publisher's bug, and
		// killing the socket over it would punish the wrong side
		c.log.Log(logger.TypeNet, logger.CodeErr, "could not encode websocket event",
			"session_id", c.SessionId, "type", e.Type, "error", err.Error())
		return nil
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.TextMessage, raw)
}
