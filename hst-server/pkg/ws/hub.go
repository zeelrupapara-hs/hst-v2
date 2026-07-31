// Package ws is the websocket fan-out: it holds the live connections and
// hands each one the events it is entitled to.
//
// The authorisation boundary is the nats subject, not a filter applied after
// delivery. A socket subscribes only to the subjects its session permits, so
// an event it may not see never reaches the process on its behalf. See
// internal/server/v1/ws.go for the subscription rules.
package ws

import (
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

const (
	// pingInterval is how often the server pings an idle socket.
	pingInterval = 25 * time.Second
	// pongWait must exceed pingInterval or every healthy client is dropped.
	pongWait = 60 * time.Second
	// writeWait bounds a single frame write, so one stuck socket cannot hold
	// its writer goroutine forever.
	writeWait = 10 * time.Second
	// maxMessageSize caps an inbound frame. Clients send keepalives, nothing
	// large, and an unbounded read is a memory exhaustion vector.
	maxMessageSize = 4096
	// egressBuffer is how far one socket may fall behind before events are
	// dropped for it.
	egressBuffer = 256
	// maxDrops is how many dropped events end the connection.
	maxDrops = 64
)

// Hub owns every live connection.
//
// Sessions are the key, but the value is a list: one login can hold the same
// session open from two tabs, and keying on the session alone would silently
// evict the first.
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]*Client
	total   int

	log *logger.Logger
}

// NewHub returns an empty hub.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients: make(map[string][]*Client),
		log:     log,
	}
}

// Add registers a connection and starts its pumps. The returned client is
// ready to receive; the caller subscribes it and then blocks on Wait.
func (h *Hub) Add(conn *websocket.Conn, snap Session, ip string) *Client {
	c := &Client{
		Id:          uuid.NewString(),
		SessionId:   snap.SessionId,
		Login:       snap.Login,
		Ip:          ip,
		Rights:      snap.ManagerRights,
		Groups:      snap.ManagerGroups,
		IsManager:   snap.IsManager,
		ConnectedAt: time.Now(),
		conn:        conn,
		hub:         h,
		log:         h.log,
		egress:      make(chan *model.Event, egressBuffer),
		closing:     make(chan struct{}),
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

// Session is what the hub needs from a session snapshot. An interface-free
// struct keeps pkg/ws independent of the cache package.
type Session struct {
	SessionId     string
	Login         int64
	IsManager     bool
	ManagerRights model.ManagerRights
	// ManagerGroups is the group access, a list of masks.
	ManagerGroups []string
}

// Remove closes one connection and forgets it. Calling it twice is safe, which
// matters because both pumps call it on their way out.
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

// CloseSession drops every connection of a session, for a revoked session or
// a forced logout.
func (h *Hub) CloseSession(sessionId string) {
	for _, c := range h.Session(sessionId) {
		c.Send(&model.Event{Type: model.EventSessionRevoked, At: time.Now().UnixNano()})
		h.Remove(sessionId, c.Id)
	}
}

// CloseLogin drops every connection belonging to one login, however many
// sessions it has open. This is what a disabled or deleted account needs.
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

// Shutdown closes every connection. Called on the way down, before nats and
// redis go, so each socket learns the server is leaving instead of timing out.
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
