package ws

import (
	"errors"

	"hstserver/model"

	"github.com/goccy/go-json"
)

// Command handles one inbound socket message.
type Command func(*Ctx) error

// Route is a command and the manager right it needs.
type Route struct {
	Handler Command
	// Right is the manager right the caller must hold; RequiresRight says whether one applies at all.
	Right         uint
	RequiresRight bool
}

// Ctx is one inbound message, the socket's equivalent of a request.
type Ctx struct {
	Client *Client
	// Type is the command name, echoed back on the reply.
	Type string
	// Id is the caller's correlation id, echoed back so a reply can be matched to its request.
	Id string
	// Data is the raw payload, parsed by BodyParser.
	Data json.RawMessage
}

// inbound is what a client sends.
type inbound struct {
	Type    string          `json:"type"`
	Id      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// reply is what a command answers with.
type reply struct {
	Type    string `json:"type"`
	Id      string `json:"id,omitempty"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// BodyParser decodes the payload into out.
func (c *Ctx) BodyParser(out any) error {
	if len(c.Data) == 0 {
		return json.Unmarshal([]byte("{}"), out)
	}
	return json.Unmarshal(c.Data, out)
}

// Login is the account that sent the command.
func (c *Ctx) Login() int64 { return c.Client.Login }

// Ip is the address the socket connected from.
func (c *Ctx) Ip() string { return c.Client.Ip }

// SendOK answers the command with its result.
func (c *Ctx) SendOK(payload any) error {
	return c.send(reply{Type: c.Type, Id: c.Id, OK: true, Payload: payload})
}

// SendError answers the command with why it failed.
func (c *Ctx) SendError(err error) error {
	if err == nil {
		return c.SendOK(nil)
	}
	return c.send(reply{Type: c.Type, Id: c.Id, OK: false, Error: err.Error()})
}

func (c *Ctx) send(r reply) error {
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// a reply is already encoded, so it rides the passthrough path
	c.Client.Send(&model.Event{Format: model.FormatText, Payload: raw})
	return nil
}

// ErrUnknownCommand is returned for a command nobody registered.
var ErrUnknownCommand = errors.New("unknown command")

// ErrNotPermitted is returned when the session lacks the right the command needs.
var ErrNotPermitted = errors.New("not permitted")
