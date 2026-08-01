package handler

import (
	"encoding/json"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// The api server builds the client envelope from the subject and these headers, so the record goes out bare.
const (
	headerFormat = "X-Format"
	headerEvent  = "X-Event"
)

// PublishWS never fails the caller: a trade already written must not be undone because a socket message could not go out.
func (h *Handler) PublishWS(subject, event string, payload any) {
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
			headerEvent:  []string{event},
		},
	}

	if err := h.Nats.NC.PublishMsg(msg); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish an event",
			"subject", subject, "error", err.Error())
	}
}

func (h *Handler) PublishResult(res *model.TradeResult) {
	if res.Login == 0 {
		return
	}

	h.PublishWS(model.SubjectAccountResult(res.Login), "trade_result", res)
}

// PublishTrade announces everything a fill changed.
func (h *Handler) PublishTrade(e *book.Entry, o *model.Order, f *Fill, a *model.Account) {
	h.PublishWS(model.SubjectAccountOrders(o.Login), "order", o)

	for _, d := range f.Deals {
		h.PublishWS(model.SubjectAccountDeals(d.Login), "deal", d)
	}

	if f.Opened != nil {
		h.PublishWS(model.SubjectAccountPositions(o.Login), "position_opened", f.Opened)
	}
	for _, p := range f.Changed {
		h.PublishWS(model.SubjectAccountPositions(o.Login), "position_changed", p)
	}
	for _, p := range f.Closed {
		h.PublishWS(model.SubjectAccountPositions(o.Login), "position_closed", p)
	}

	h.PublishWS(model.SubjectAccountSummary(o.Login), "account", a)
}

// PublishAccount announces a new money state on its own, for the tick path where nothing traded.
func (h *Handler) PublishAccount(a *model.Account) {
	h.PublishWS(model.SubjectAccountSummary(a.Login), "account", a)
}

// PublishDealing offers one request to one dealer's queue.
func (h *Handler) PublishDealing(dealer int64, e *model.DealingEvent) {
	event := "dealer_request"
	if e.EventType != model.DealingEventOffer {
		event = "dealer_request_done"
	}

	h.PublishWS(model.SubjectDealerRequests(dealer), event, e)
}
