// Package ws is the websocket fan-out: it holds the live connections and hands each one the events it is entitled to.
package ws

import (
	"errors"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

const (
	// pingInterval is how often the server pings an idle socket.
	pingInterval = 25 * time.Second
	// pongWait must exceed pingInterval or every healthy client is dropped.
	pongWait = 60 * time.Second
	// writeWait bounds a single frame write, so one stuck socket cannot hold its writer goroutine forever.
	writeWait = 10 * time.Second
	// maxMessageSize caps an inbound frame.
	maxMessageSize = 4096
	// egressBuffer is how far one socket may fall behind before events are dropped for it.
	egressBuffer = 256
	// maxDrops is how many dropped events end the connection.
	maxDrops = 64
)

// Handler answers one inbound frame.
type Handler func(*Ctx) error

// ErrorHandler turns a handler's error into what the client is sent.
type ErrorHandler func(err error) *model.Event

// ErrUnknownEvent is what an unregistered event type turns into.
var ErrUnknownEvent = errors.New("unknown event type")

// Hub owns every live connection.
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]*Client
	total   int

	// RouterMap binds an inbound event type to its handler, written once at startup.
	RouterMap map[model.EventType]Handler
	onError   ErrorHandler

	log *logger.Logger
}

// NewHub returns an empty hub.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:   make(map[string][]*Client),
		RouterMap: make(map[model.EventType]Handler),
		onError:   defaultErrorHandler,
		log:       log,
	}
}

// RegisterRoute binds an inbound event type to a handler.
func (h *Hub) RegisterRoute(event model.EventType, handler Handler) { h.RouterMap[event] = handler }

// SetErrorHandler decides what a handler's error turns into on the wire.
func (h *Hub) SetErrorHandler(cb ErrorHandler) { h.onError = cb }

// Dispatch parses a frame and calls the bound handler.
func (h *Hub) Dispatch(c *Client, data []byte) {
	var in struct {
		Type    model.EventType `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		c.Send(h.onError(err))
		return
	}

	handler, ok := h.RouterMap[in.Type]
	if !ok {
		c.Send(h.onError(ErrUnknownEvent))
		return
	}

	if err := handler(NewCtx(c, in.Type, in.Payload)); err != nil {
		c.Send(h.onError(err))
	}
}

func defaultErrorHandler(error) *model.Event {
	return &model.Event{Type: model.EventInternalServerError}
}

// Add registers a connection and starts its pumps.
func (h *Hub) Add(conn *websocket.Conn, snap Session, ip, os string) *Client {
	c := &Client{
		Id:             uuid.NewString(),
		SessionId:      snap.SessionId,
		Login:          snap.Login,
		Scope:          snap.Scope,
		Ip:             ip,
		rights:         snap.ManagerRights,
		groups:         snap.ManagerGroups,
		IsManager:      snap.IsManager,
		ConnectionType: snap.ConnectionType,
		Os:             os,
		ConnectedAt:    time.Now(),
		conn:           conn,
		hub:            h,
		log:            h.log,
		egress:         make(chan *model.Event, egressBuffer),
		closing:        make(chan struct{}),
	}

	h.mu.Lock()
	h.clients[c.SessionId] = append(h.clients[c.SessionId], c)
	h.total++
	total := h.total
	h.mu.Unlock()

	go c.writePump()
	go c.readPump()

	h.log.Log(logger.TypeNet, logger.CodeOK, "websocket connected",
		"session_id", c.SessionId, "login", c.Login, "ip", ip, "connections", total)

	return c
}

// Session is what the hub needs from a session snapshot.
type Session struct {
	SessionId string
	Login     int64
	// Scope is the password slot the session authenticated with: investor may look and not touch.
	Scope int32
	// ConnectionType is the terminal that authenticated.
	ConnectionType int32
	IsManager      bool
	ManagerRights  model.ManagerRights
	// ManagerGroups is the group access, a list of masks.
	ManagerGroups []string
}

// Remove closes one connection and forgets it.
func (h *Hub) Remove(sessionId, clientId string) {
	h.mu.Lock()
	list := h.clients[sessionId]
	var found *Client
	for i := range list {
		if list[i].Id == clientId {
			found = list[i]
			h.clients[sessionId] = append(list[:i], list[i+1:]...)
			if len(h.clients[sessionId]) == 0 {
				delete(h.clients, sessionId)
			}
			h.total--
			break
		}
	}
	total := h.total
	h.mu.Unlock()

	if found == nil {
		return
	}

	found.close()

	h.log.Log(logger.TypeNet, logger.CodeOK, "websocket disconnected",
		"session_id", sessionId, "login", found.Login,
		"dropped", found.Dropped(), "connections", total)
}

// Session returns the connections of one session.
func (h *Hub) Session(sessionId string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := h.clients[sessionId]
	out := make([]*Client, len(list))
	copy(out, list)
	return out
}

// CloseSession drops every connection of a session, for a revoked session or a forced logout.
func (h *Hub) CloseSession(sessionId string) {
	for _, c := range h.Session(sessionId) {
		c.Send(&model.Event{Type: model.EventSessionRevoked, At: time.Now().UnixNano()})
		h.Remove(sessionId, c.Id)
	}
}

// Login returns every connection belonging to one login, for a refresh that has to rebuild what they listen to.
func (h *Hub) Login(login int64) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var out []*Client
	for _, list := range h.clients {
		for _, c := range list {
			if c.Login == login {
				out = append(out, c)
			}
		}
	}
	return out
}

// CloseLogin drops every connection belonging to one login, however many sessions it has open.
func (h *Hub) CloseLogin(login int64) {
	h.mu.RLock()
	var ids []string
	for sid, list := range h.clients {
		for i := range list {
			if list[i].Login == login {
				ids = append(ids, sid)
				break
			}
		}
	}
	h.mu.RUnlock()

	for _, sid := range ids {
		h.CloseSession(sid)
	}
}

// Count is the number of live connections.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.total
}

// Sessions is the number of distinct sessions holding a connection.
func (h *Hub) Sessions() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Shutdown closes every connection.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	all := make([]*Client, 0, h.total)
	for _, list := range h.clients {
		all = append(all, list...)
	}
	h.clients = make(map[string][]*Client)
	h.total = 0
	h.mu.Unlock()

	for _, c := range all {
		c.close()
	}

	h.log.Log(logger.TypeNet, logger.CodeOK, "websocket hub stopped", "closed", len(all))
}
