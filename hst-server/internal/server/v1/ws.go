package v1

import (
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/ws"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	natscore "github.com/nats-io/nats.go"
)

// rightSubjects maps a manager right to the subject a holder of it listens on.
//
// This table is the authorisation model of the websocket layer. A socket
// subscribes to a right's subject only when the session actually holds that
// right, so an event it may not see is never delivered to it in the first
// place. There is no filtering after the fact to get wrong.
//
// Publishing to one of these is how a service reaches "every manager who may
// see this", without knowing who is connected.
var rightSubjects = []struct {
	right uint
	name  string
}{
	{model.MgrRightSrvJournals, "journals"},
	{model.MgrRightAccRead, "users"},
	{model.MgrRightClientsAccess, "clients"},
	{model.MgrRightCfgGroups, "groups"},
	{model.MgrRightCfgSymbols, "symbols"},
	{model.MgrRightCfgManagers, "managers"},
}

// ServeWS upgrades an authenticated request and attaches it to the hub.
//
// Authentication has already happened: UpgradeWS finds the token and Protect
// validates it, exactly as it does for every other route, so a socket cannot
// exist without a live session behind it.
func (s *HttpServer) ServeWS(c *websocket.Conn) {
	snap, ok := c.Locals(nethttp.LocalsClient).(*cache.Snapshot)
	if !ok || snap == nil {
		// Protect guarantees this, so reaching it means the chain was changed
		_ = c.Close()
		return
	}

	ip, _ := c.Locals(nethttp.LocalsIp).(string)

	client := s.Hub.Add(c, ws.Session{
		SessionId:     snap.SessionId,
		Login:         snap.Login,
		IsManager:     snap.IsManager,
		ManagerRights: snap.ManagerRights,
	}, ip)

	if err := s.subscribe(client); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeErr, "could not subscribe a websocket",
			"session_id", client.SessionId, "error", err.Error())
		s.Hub.Remove(client.SessionId, client.Id)
		return
	}

	client.Send(&model.Event{Type: model.EventWelcome, At: time.Now().UnixNano()})

	// the socket is closed as soon as this returns, so park here until the
	// connection is torn down
	client.Wait()
}

// subscribe opens every nats subscription this socket is entitled to. A
// failure part way through is not left half wired: the caller removes the
// client, which unsubscribes whatever was already opened.
func (s *HttpServer) subscribe(c *ws.Client) error {
	// its own session, and every session of its login
	if err := s.subscribeTo(c, model.SubjectSession(c.SessionId)); err != nil {
		return err
	}
	if err := s.subscribeTo(c, model.SubjectLogin(c.Login)); err != nil {
		return err
	}
	if err := s.subscribeTo(c, model.SubjectBroadcast()); err != nil {
		return err
	}

	// and one subject per right it holds
	if c.IsManager {
		for _, r := range rightSubjects {
			if !c.Rights.Has(r.right) {
				continue
			}
			if err := s.subscribeTo(c, model.SubjectRight(r.name)); err != nil {
				return err
			}
		}
	}

	return nil
}

// subscribeTo wires one subject into one socket.
func (s *HttpServer) subscribeTo(c *ws.Client, subject string) error {
	sub, err := s.Nats.NC.Subscribe(subject, func(msg *natscore.Msg) {
		c.Send(eventFromMsg(msg))
	})
	if err != nil {
		return err
	}

	c.AddSub(sub)
	return nil
}

// eventFromMsg turns a nats message into the event the client receives.
//
// The event name comes from the subject, so the common case needs no envelope
// and a publisher cannot label an event as something it was not routed as.
//
// The format comes from the publisher: any service can put json, packed
// binary or plain text on a subject and it reaches the socket in the matching
// frame, untouched. Declare it with the X-Format or Content-Type header;
// without one the payload is sniffed, so a publisher already sending json
// needs no changes at all.
func eventFromMsg(msg *natscore.Msg) *model.Event {
	format, declared := model.FormatFromHeader(
		msg.Header.Get(model.HeaderFormat),
		msg.Header.Get(model.HeaderContentType),
	)
	if !declared {
		format = model.SniffFormat(msg.Data)
	}

	e := &model.Event{
		Type:    model.EventTypeFromSubject(msg.Subject),
		Payload: msg.Data,
		Format:  format,
		At:      time.Now().UnixNano(),
	}
	if e.Type == "" {
		e.Type = model.EventError
	}
	return e
}

// PublishWS sends a json event to a subject, for any handler that wants to
// notify the connected clients.
func (s *HttpServer) PublishWS(subject string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.PublishWSRaw(subject, raw, model.FormatJSON)
}

// PublishWSRaw sends bytes that are already encoded, in the format given. This
// is the path a tick stream or any other packed feed takes: nothing here parses
// or copies the payload, and the client receives the same bytes in a binary
// frame.
func (s *HttpServer) PublishWSRaw(subject string, payload []byte, format model.Format) error {
	msg := &natscore.Msg{
		Subject: subject,
		Data:    payload,
		Header:  natscore.Header{},
	}

	switch format {
	case model.FormatBinary:
		msg.Header.Set(model.HeaderFormat, "binary")
	case model.FormatText:
		msg.Header.Set(model.HeaderFormat, "text")
	default:
		msg.Header.Set(model.HeaderFormat, "json")
	}

	return s.Nats.NC.PublishMsg(msg)
}

// WSStats is the websocket hub report.
type WSStats struct {
	// Connections is every live socket.
	Connections int `json:"connections"`
	// Sessions is how many distinct sessions hold one, which is lower whenever
	// someone has two tabs open.
	Sessions int `json:"sessions"`
}

// WSStats reports the live websocket connections.
//
//	@Id			WSStats
//	@Tags		System
//	@Produce	json
//	@Success	200	{object}	Response{data=WSStats}
//	@Failure	401	{object}	Response
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/system/monitor/ws [get]
func (s *HttpServer) WSStats(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, WSStats{
		Connections: s.Hub.Count(),
		Sessions:    s.Hub.Sessions(),
	})
}
