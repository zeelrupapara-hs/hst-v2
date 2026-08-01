package v1

import (
	"context"
	"time"

	"hstserver/model"
	"hstserver/pkg/cache"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/logger"
	"hstserver/pkg/ws"
	"hstserver/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	natscore "github.com/nats-io/nats.go"
)

// rightSubjects are the records with no group of their own: access to them is a right, not a path.
var rightSubjects = []struct {
	right   uint
	subject string
}{
	{model.MgrRightCfgSymbols, model.SubjectSymbol},
	{model.MgrRightCfgHolidays, model.SubjectHoliday},
	{model.MgrRightCfgGroups, model.SubjectLeverage},
	{model.MgrRightCfgManagers, model.SubjectManager},
}

// ServeWS upgrades an authenticated request and attaches it to the hub.
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
		Scope:         snap.Scope,
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

	// the socket is closed as soon as this returns, so park here until the connection is torn down
	client.Wait()

	// a terminal that has gone away is not watching prices any more
	dropMarketWatcher(client.SessionId)
}

// subscribe opens every nats subscription this socket is entitled to.
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

// subjectsFor is everything one socket listens on, given what its session holds.
func (s *HttpServer) subjectsFor(c *ws.Client, rights model.ManagerRights, groups []string) []string {
	// its own session, every session of its login, and the whole floor
	subjects := []string{
		model.SubjectSession(c.SessionId),
		model.SubjectLogin(c.Login),
		model.SubjectBroadcast(),
	}

	// a trading account hears about itself and nothing else: it holds no rights and covers no tree
	if !c.IsManager {
		return append(subjects, model.SubjectTrader(c.Login))
	}

	// the journal is the staff record of who did what
	subjects = append(subjects, model.SubjectJournal(c.Login))

	// a dealer works its own queue of requests waiting on an answer
	if rights.Has(model.MgrRightTradesDealer) {
		subjects = append(subjects, model.SubjectDealerRequests(c.Login))
	}

	// one subject per right, for records that have no group: symbols, other managers, the server journal
	for _, r := range rightSubjects {
		if rights.Has(r.right) {
			subjects = append(subjects, r.subject)
		}
	}

	// and one per family it may see, crossed with each group mask it holds.
	return append(subjects, model.Subscriptions(rights, groups)...)
}

// openSubs subscribes to every subject the access allows.
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

// RefreshLogin rebuilds the subscriptions of every socket of one login, after its access was rewritten.
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
func eventFromMsg(msg *natscore.Msg) *model.Event {
	format, declared := model.FormatFromHeader(
		msg.Header.Get(model.HeaderFormat),
		msg.Header.Get(model.HeaderContentType),
	)
	if !declared {
		format = model.SniffFormat(msg.Data)
	}

	typ, group := model.ParseSubject(msg.Subject)

	// a record change names what happened in a header, since the subject only says who may see it
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

// NotifySystem publishes a record change for the other services.
func (s *HttpServer) NotifySystem(subject string, payload any) {
	if err := s.PublishWS(subject, payload); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish a system event",
			"subject", subject, "error", err.Error())
	}
}

// write journal entry
func (s *HttpServer) JournalEntry(c *fiber.Ctx, code logger.Code, message string, detail any) {
	snap, _ := utils.GetClient(c)

	var login int64
	if snap != nil {
		login = snap.Login
	}

	s.WriteJournal(c.UserContext(), login, utils.GetRealIP(c), code, message, detail)
}

// WriteJournal records a line for an actor that is not a session, such as a public signup.
func (s *HttpServer) WriteJournal(ctx context.Context, login int64, ip string, code logger.Code, message string, detail any) {

	raw, err := json.Marshal(detail)
	if err != nil {
		// a detail that cannot be encoded is the caller's bug, and losing the whole entry over it would hide what actually happened
		s.Log.Log(logger.TypeSys, logger.CodeWarn, "could not encode a journal detail",
			"message", message, "error", err.Error())
		raw = nil
	}

	if err := s.Journal.Entry(ctx, &model.Journal{
		Type: int32(logger.TypeCfg),
		// #nosec G115 -- a logger code is 0..4, far inside the column
		Code:    int32(code),
		Login:   login,
		Ip:      ip,
		Message: message,
		Detail:  raw,
	}); err != nil {
		s.Log.Log(logger.TypeSys, logger.CodeWarn, "could not write a journal entry",
			"message", message, "error", err.Error())
	}
}

// PublishWS sends a json event to a subject, for any handler that wants to notify the connected clients.
func (s *HttpServer) PublishWS(subject string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.PublishWSRaw(subject, raw, model.FormatJSON)
}

// PublishWSRaw sends bytes that are already encoded, in the format given.
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
	// Sessions is how many distinct sessions hold one, which is lower whenever someone has two tabs open.
	Sessions int `json:"sessions"`
}

// WSStats reports the live websocket connections.
func (s *HttpServer) WSStats(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, WSStats{
		Connections: s.Hub.Count(),
		Sessions:    s.Hub.Sessions(),
	})
}

// ViewGroupRef identifies a record that no longer exists, for a delete event.
type ViewGroupRef struct {
	GroupID int    `json:"group_id"`
	Group   string `json:"group"`
}

// the event type comes from a header, since the subject only says who may see it
type ViewUserRef struct {
	Login int64  `json:"login"`
	Group string `json:"group"`
}

// NotifyWS publishes a record change on the subject the caller named.
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

// ViewManagerRef identifies a manager that no longer exists.
type ViewManagerRef struct {
	Login int64 `json:"login"`
}

// ViewCommissionRef identifies a commission that no longer exists.
type ViewCommissionRef struct {
	GroupID      int `json:"group_id"`
	CommissionID int `json:"commission_id"`
}

// ViewLeverageRef identifies a leverage profile that no longer exists.
type ViewLeverageRef struct {
	LeverageId int `json:"leverage_id"`
}

// ViewHolidayRef identifies a holiday that no longer exists.
type ViewHolidayRef struct {
	HolidayId int `json:"holiday_id"`
}

// ViewSymbolRef identifies a symbol that no longer exists.
type ViewSymbolRef struct {
	SymbolId int    `json:"symbol_id"`
	Symbol   string `json:"symbol"`
	Path     string `json:"path"`
}

// ViewClientRef identifies a client that no longer exists.
type ViewClientRef struct {
	ClientId int64 `json:"client_id"`
}

// ClientGroups is the distinct set of groups a client is present in, through the logins it owns.
func (s *HttpServer) ClientGroups(ctx context.Context, clientId int64) []string {
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

// NotifyClient announces a client change to every group the client has a login in.
func (s *HttpServer) NotifyClient(ctx context.Context, clientId int64, event string, payload any) {
	s.NotifyClientIn(s.ClientGroups(ctx, clientId), event, payload)
}

// NotifyClientIn is NotifyClient with the groups already read, for a delete where the logins are detached before the row goes.
func (s *HttpServer) NotifyClientIn(groups []string, event string, payload any) {
	// a client with no login yet sits under root, where only a manager with unrestricted access is listening
	if len(groups) == 0 {
		s.NotifyWS(model.SubjectClient(""), event, payload)
		return
	}

	for _, g := range groups {
		s.NotifyWS(model.SubjectClient(g), event, payload)
	}
}
