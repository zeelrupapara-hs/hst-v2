package v1

import (
	"context"
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
		ManagerGroups: snap.ManagerGroups,
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
	rights, groups := c.Access()

	subs, err := s.openSubs(c, rights, groups)
	if err != nil {
		closeSubs(subs)
		return err
	}

	c.ReplaceSubs(rights, groups, subs)
	return nil
}

// subjectsFor is everything one socket listens on, given what its session
// holds. Building the list separately from opening it lets a refresh compare
// the two without touching the connection.
func (s *HttpServer) subjectsFor(c *ws.Client, rights model.ManagerRights, groups []string) []string {
	// its own session, every session of its login, and the whole floor
	subjects := []string{
		model.SubjectSession(c.SessionId),
		model.SubjectLogin(c.Login),
		model.SubjectBroadcast(),
	}

	if !c.IsManager {
		return subjects
	}

	// one subject per right, for records that have no group: symbols, other
	// managers, the server journal
	for _, r := range rightSubjects {
		if rights.Has(r.right) {
			subjects = append(subjects, model.SubjectRight(r.name))
		}
	}

	// and one per family it may see, crossed with each group mask it holds.
	// The mask is the subscription, so a group created later under a granted
	// path is covered without anyone re-subscribing.
	return append(subjects, model.Subscriptions(rights, groups)...)
}

// openSubs subscribes to every subject the access allows. On a failure part way
// through it returns what it managed to open, for the caller to close.
func (s *HttpServer) openSubs(c *ws.Client, rights model.ManagerRights, groups []string) ([]ws.Unsubscriber, error) {
	subjects := s.subjectsFor(c, rights, groups)

	subs := make([]ws.Unsubscriber, 0, len(subjects))
	for _, subject := range subjects {
		sub, err := s.Nats.NC.Subscribe(subject, func(msg *natscore.Msg) {
			c.Send(eventFromMsg(msg))
		})
		if err != nil {
			return subs, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func closeSubs(subs []ws.Unsubscriber) {
	for _, sub := range subs {
		_ = sub.Unsubscribe()
	}
}

// RefreshLogin rebuilds the subscriptions of every socket of one login, after
// its access was rewritten.
//
// The connection survives. A manager that gains a group starts receiving that
// group's events on the socket it already had, which is the whole point of
// refreshing a session rather than revoking it.
func (s *HttpServer) RefreshLogin(login int64) {
	for _, c := range s.Hub.Login(login) {
		snap, err := s.OAuth2.LoadSnapshot(context.Background(), c.SessionId)
		if err != nil {
			// the session is gone rather than changed, so the socket goes too
			s.Hub.Remove(c.SessionId, c.Id)
			continue
		}

		subs, err := s.openSubs(c, snap.ManagerRights, snap.ManagerGroups)
		if err != nil {
			closeSubs(subs)
			s.Log.Log(logger.TypeNet, logger.CodeErr, "could not resubscribe a websocket",
				"session_id", c.SessionId, "error", err.Error())
			continue
		}

		c.ReplaceSubs(snap.ManagerRights, snap.ManagerGroups, subs)

		s.Log.Log(logger.TypeNet, logger.CodeOK, "websocket access refreshed",
			"session_id", c.SessionId, "login", login, "subjects", len(subs))
	}
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

	typ, group := model.ParseSubject(msg.Subject)

	// a record change names what happened in a header, since the subject only
	// says who may see it
	if event := msg.Header.Get(model.HeaderEvent); event != "" {
		typ = event
	}

	e := &model.Event{
		Type:    typ,
		Group:   group,
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

// ViewGroupRef identifies a record that no longer exists, for a delete event.
// The reader only needs to know which row to drop.
type ViewGroupRef struct {
	GroupID int    `json:"group_id"`
	Group   string `json:"group"`
}

// ViewUserRef identifies a login for a delete or a move, where the whole
// record is either gone or no longer this audience's business.
type ViewUserRef struct {
	Login int64  `json:"login"`
	Group string `json:"group"`
}

// NotifyWS publishes a record change on the subject the caller named.
//
// The publisher says what changed and where it lives, and nothing about who may
// see it: the group path inside the subject is the authorisation, matched by
// nats against each manager's access masks. No lookup of connected clients, no
// query of who holds which right, one publish however many are listening.
//
// A failure is logged and swallowed. The write already succeeded; failing the
// request because a notification did not go out would be the wrong trade.
func (s *HttpServer) NotifyWS(subject, event string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeErr, "could not encode a websocket event",
			"subject", subject, "event", event, "error", err.Error())
		return
	}

	msg := &natscore.Msg{
		Subject: subject,
		Data:    raw,
		Header: natscore.Header{
			model.HeaderFormat: []string{"json"},
			model.HeaderEvent:  []string{event},
		},
	}

	if err := s.Nats.NC.PublishMsg(msg); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish a websocket event",
			"subject", subject, "event", event, "error", err.Error())
	}
}

// ViewGroupSymbolRef identifies an override that no longer exists.
type ViewGroupSymbolRef struct {
	GroupID  int `json:"group_id"`
	SymbolID int `json:"symbol_id"`
}

// ViewClientRef identifies a client that no longer exists.
type ViewClientRef struct {
	ClientId int64 `json:"client_id"`
}

// clientGroups is the distinct set of groups a client is present in, through
// the logins it owns.
//
// A client has no group column of its own; it is a person, and the person can
// hold a demo login in one tree and a live login in another. Its audience is
// therefore the union of those trees.
func (s *HttpServer) clientGroups(ctx context.Context, clientId int64) []string {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT DISTINCT "group" FROM hst.users WHERE client_id = $1`, clientId)
	if err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "could not read a client's groups",
			"client_id", clientId, "error", err.Error())
		return nil
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return out
		}
		out = append(out, g)
	}
	return out
}

// notifyClient announces a client change to every group the client has a login
// in. A manager covering two of those groups receives it twice, which is
// harmless: the event carries the whole record, so applying it twice is the
// same as applying it once.
func (s *HttpServer) notifyClient(ctx context.Context, clientId int64, event string, payload any) {
	s.notifyClientIn(s.clientGroups(ctx, clientId), event, payload)
}

// notifyClientIn is notifyClient with the groups already read, for a delete
// where the logins are detached before the row goes.
func (s *HttpServer) notifyClientIn(groups []string, event string, payload any) {
	// a client with no login yet sits under root, where only a manager with
	// unrestricted access is listening
	if len(groups) == 0 {
		s.NotifyWS(model.SubjectClient(""), event, payload)
		return
	}

	for _, g := range groups {
		s.NotifyWS(model.SubjectClient(g), event, payload)
	}
}
