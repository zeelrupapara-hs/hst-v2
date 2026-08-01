package handler

import (
	"encoding/json"
	"fmt"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// Telling everyone what happened.
//
// The engine owns the trading state, so it is the only thing that can say a position opened. It
// publishes to the same subjects the API server already fans out to its websockets: a trading
// account hears about itself on ws.t.<login>, and managers hear group-scoped events.

// The API server builds the envelope its clients see from the subject and these headers, so
// the engine publishes the record itself and names the event beside it. Wrapping it here as
// well would reach the client wrapped twice.
const (
	headerFormat = "X-Format"
	headerEvent  = "X-Event"
)

// Subjects the engine publishes on.
func subjectAccount(login int64) string   { return fmt.Sprintf("ws.t.%d.account", login) }
func subjectPositions(login int64) string { return fmt.Sprintf("ws.t.%d.positions", login) }
func subjectOrders(login int64) string    { return fmt.Sprintf("ws.t.%d.orders", login) }
func subjectDeals(login int64) string     { return fmt.Sprintf("ws.t.%d.deals", login) }
func subjectResult(login int64) string    { return fmt.Sprintf("ws.t.%d.result", login) }

// publish sends one event, and never fails the caller: a trade that is already written must not
// be undone because a socket message could not go out.
func (h *Handler) publish(subject, kind string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not encode an event",
			"subject", subject, "error", err.Error())
		return
	}

	msg := &natscore.Msg{
		Subject: subject,
		Data:    body,
		Header: natscore.Header{
			headerFormat: []string{"json"},
			headerEvent:  []string{kind},
		},
	}

	if err := h.Nats.NC.PublishMsg(msg); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish an event",
			"subject", subject, "error", err.Error())
	}
}

// publishResult tells the client what became of its request.
func (h *Handler) publishResult(res *TradeResult) {
	if res.Login == 0 {
		return
	}

	h.publish(subjectResult(res.Login), "trade_result", res)
}

// publishTrade announces everything a fill changed.
func (h *Handler) publishTrade(e *book.Entry, o *model.Order, f *Fill, a *model.Account) {
	h.publish(subjectOrders(o.Login), "order", o)

	for _, d := range f.Deals {
		h.publish(subjectDeals(d.Login), "deal", d)
	}

	if f.Opened != nil {
		h.publish(subjectPositions(o.Login), "position_opened", f.Opened)
	}
	for _, p := range f.Changed {
		h.publish(subjectPositions(o.Login), "position_changed", p)
	}
	for _, p := range f.Closed {
		h.publish(subjectPositions(o.Login), "position_closed", p)
	}

	h.publish(subjectAccount(o.Login), "account", a)
}

// publishAccount announces a new money state on its own, for the tick path where nothing traded
// but the numbers moved.
func (h *Handler) publishAccount(a *model.Account) {
	h.publish(subjectAccount(a.Login), "account", a)
}
