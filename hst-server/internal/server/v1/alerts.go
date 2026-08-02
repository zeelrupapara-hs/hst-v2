package v1

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"hstserver/model"
	"hstserver/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// What an alert watches.
const (
	AlertKindAsk         = 1
	AlertKindBid         = 2
	AlertKindBalance     = 3
	AlertKindEquity      = 4
	AlertKindMarginLevel = 5
)

// How the watched value is compared with the alert's own.
const (
	AlertCondGreater = 1
	AlertCondLess    = 2
)

// AlertIsPrice is whether a kind is quoted rather than banked.
func AlertIsPrice(kind int32) bool { return kind == AlertKindAsk || kind == AlertKindBid }

// AlertKindValid is whether a kind is one of the five the product defines.
func AlertKindValid(kind int32) bool { return kind >= AlertKindAsk && kind <= AlertKindMarginLevel }

// AlertCondValid is whether a condition is one of the two the product defines.
func AlertCondValid(cond int32) bool { return cond == AlertCondGreater || cond == AlertCondLess }

// armedAlert is one enabled alert, as the evaluator holds it.
type armedAlert struct {
	AlertId int64
	Login   int64
	Symbol  string
	Kind    int32
	Cond    int32
	Value   float64
}

// alertBook is every armed alert, indexed the two ways it is looked up.
type alertBook struct {
	mu       sync.RWMutex
	bySymbol map[string][]armedAlert
	byLogin  map[int64][]armedAlert
}

var alerts = &alertBook{bySymbol: map[string][]armedAlert{}, byLogin: map[int64][]armedAlert{}}

// StartAlerts loads the armed alerts and listens for the account summaries that can fire them.
func (s *HttpServer) StartAlerts() {
	if err := s.ReloadAlerts(context.Background()); err != nil {
		s.Log.Log(logger.TypeSys, logger.CodeErr, "could not load alerts", "error", err.Error())
	}

	if _, err := s.Nats.NC.Subscribe("ws.t.*.summary", s.AlertSummaryHandler); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeErr, "could not watch account summaries for alerts",
			"error", err.Error())
	}
}

// ReloadAlerts rebuilds the in memory book from the table, cheap while alerts number in the thousands.
func (s *HttpServer) ReloadAlerts(ctx context.Context) error {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT alert_id, login, COALESCE(symbol, ''), kind, condition, value
		   FROM hst.alerts WHERE enabled`)
	if err != nil {
		return err
	}
	defer rows.Close()

	bySymbol := map[string][]armedAlert{}
	byLogin := map[int64][]armedAlert{}

	for rows.Next() {
		var a armedAlert
		if err := rows.Scan(&a.AlertId, &a.Login, &a.Symbol, &a.Kind, &a.Cond, &a.Value); err != nil {
			return err
		}
		if AlertIsPrice(a.Kind) {
			bySymbol[a.Symbol] = append(bySymbol[a.Symbol], a)
			continue
		}
		byLogin[a.Login] = append(byLogin[a.Login], a)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	alerts.mu.Lock()
	alerts.bySymbol, alerts.byLogin = bySymbol, byLogin
	alerts.mu.Unlock()

	return nil
}

// AlertsOnTick fires the price alerts standing on one instrument.
func (s *HttpServer) AlertsOnTick(t *model.Tick) {
	alerts.mu.RLock()
	armed := alerts.bySymbol[t.Symbol]
	alerts.mu.RUnlock()

	for _, a := range armed {
		price := t.Bid
		if a.Kind == AlertKindAsk {
			price = t.Ask
		}
		s.fireAlert(a, price)
	}
}

// AlertSummaryHandler fires the money alerts standing on one account.
func (s *HttpServer) AlertSummaryHandler(msg *natscore.Msg) {
	// summary,login,balance,credit,equity,margin,margin_free,margin_level%,profit,positions...
	f := strings.Split(string(msg.Data), ",")
	if len(f) < 8 {
		return
	}

	login, err := strconv.ParseInt(f[1], 10, 64)
	if err != nil {
		return
	}

	alerts.mu.RLock()
	armed := alerts.byLogin[login]
	alerts.mu.RUnlock()
	if len(armed) == 0 {
		return
	}

	at := map[int32]string{AlertKindBalance: f[2], AlertKindEquity: f[4],
		AlertKindMarginLevel: strings.TrimSuffix(f[7], "%")}

	for _, a := range armed {
		v, err := strconv.ParseFloat(at[a.Kind], 64)
		if err != nil {
			continue
		}
		s.fireAlert(a, v)
	}
}

// fireAlert disarms an alert whose level has been reached and tells the terminal.
func (s *HttpServer) fireAlert(a armedAlert, value float64) {
	if a.Cond == AlertCondGreater && value <= a.Value {
		return
	}
	if a.Cond == AlertCondLess && value >= a.Value {
		return
	}

	now := time.Now().UnixNano()

	// the update is conditional, so two ticks racing on the same alert only fire it once
	tag, err := s.DB.DB.Exec(context.Background(),
		`UPDATE hst.alerts SET enabled = FALSE, triggered_at = $2, updated_at = $2
		  WHERE alert_id = $1 AND enabled`, a.AlertId, now)
	if err != nil || tag.RowsAffected() == 0 {
		return
	}

	s.dropArmedAlert(a)

	payload, err := json.Marshal(map[string]any{"alert_id": a.AlertId, "login": a.Login,
		"symbol": a.Symbol, "kind": a.Kind, "condition": a.Cond, "value": a.Value,
		"price": value, "triggered_at": now})
	if err != nil {
		return
	}

	for _, c := range s.Hub.Login(a.Login) {
		c.Send(&model.Event{Type: model.EventAlertTriggered, Format: model.FormatJSON,
			Payload: payload, At: now})
	}
}

// dropArmedAlert forgets one alert that has fired, without reloading the book.
func (s *HttpServer) dropArmedAlert(a armedAlert) {
	alerts.mu.Lock()
	defer alerts.mu.Unlock()

	if AlertIsPrice(a.Kind) {
		alerts.bySymbol[a.Symbol] = withoutAlert(alerts.bySymbol[a.Symbol], a.AlertId)
		return
	}

	alerts.byLogin[a.Login] = withoutAlert(alerts.byLogin[a.Login], a.AlertId)
}

func withoutAlert(list []armedAlert, alertId int64) []armedAlert {
	out := make([]armedAlert, 0, len(list))
	for _, a := range list {
		if a.AlertId != alertId {
			out = append(out, a)
		}
	}

	return out
}
