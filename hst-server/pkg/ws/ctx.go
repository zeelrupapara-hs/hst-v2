package ws

import (
	"errors"

	"hstserver/model"

	"github.com/goccy/go-json"
)

// ErrConnectionLost is returned when the socket went away before the answer could be queued.
var ErrConnectionLost = errors.New("connection is lost")

// Ctx is one inbound frame, ready for a handler.
type Ctx struct {
	Client *Client
	Type   model.EventType
	Data   []byte
	Event  *model.Event
}

// NewCtx wraps one inbound frame.
func NewCtx(c *Client, typ model.EventType, data []byte) *Ctx {
	return &Ctx{
		Client: c,
		Type:   typ,
		Data:   data,
		Event:  &model.Event{Type: typ, Payload: data},
	}
}

// BodyParser decodes the frame's payload.
func (c *Ctx) BodyParser(out any) error { return json.Unmarshal(c.Data, out) }

// SendEvent queues one event back on this socket.
func (c *Ctx) SendEvent(e *model.Event) error {
	if c.Client == nil || !c.Client.Send(e) {
		return ErrConnectionLost
	}
	return nil
}

// Login is the authenticated login, never the one in the payload.
func (c *Ctx) Login() int64 { return c.Client.Login }

// Scope is the password slot the session authenticated with.
func (c *Ctx) Scope() int32 { return c.Client.Scope }

// IsManager says which side of the plain and My split the caller is on.
func (c *Ctx) IsManager() bool { return c.Client.IsManager }

// Rights is what the session holds, for the dealer routes.
func (c *Ctx) Rights() model.ManagerRights { return c.Client.Rights() }
