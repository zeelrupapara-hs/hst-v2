package handler

import (
	"encoding/json"
	"time"

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
func (h *Handler) PublishWS(subject string, event model.EventType, payload any) {
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
			headerEvent:  []string{string(event)},
		},
	}

	if err := h.Nats.NC.PublishMsg(msg); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish an event",
			"subject", subject, "error", err.Error())
	}
}

// PublishText sends an already formatted payload, for messages too frequent to afford an envelope.
func (h *Handler) PublishText(subject string, event model.EventType, payload string) {
	msg := &natscore.Msg{
		Subject: subject,
		Data:    []byte(payload),
		Header: natscore.Header{
			headerFormat: []string{"text"},
			headerEvent:  []string{string(event)},
		},
	}

	if err := h.Nats.NC.PublishMsg(msg); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not publish an event",
			"subject", subject, "error", err.Error())
	}
}

// PublishRejected tells an account its request was refused, which is the only way a caller that
// did not wait for a reply learns the outcome.
func (h *Handler) PublishRejected(res *model.TradeResult) {
	if res.Login == 0 || res.RetCode == int32(model.RetOK) {
		return
	}

	h.PublishWS(model.SubjectAccountOrders(res.Login), model.EventOrderRejected, res)
}

// PublishPosition announces one position event to its account, and to the managers whose
// group masks cover the account's group. An empty group skips the manager copy rather than
// leaking the event to the root subject.
func (h *Handler) PublishPosition(group string, event model.EventType, p *model.Position) {
	wire := model.NewWirePosition(p)
	h.PublishWS(model.SubjectAccountPositions(p.Login), event, wire)
	if group != "" {
		h.PublishWS(model.SubjectGroupPositions(group), event, wire)
	}
}

// PublishOrder is PublishPosition for an order.
func (h *Handler) PublishOrder(group string, event model.EventType, o *model.Order) {
	wire := model.NewWireOrder(o)
	h.PublishWS(model.SubjectAccountOrders(o.Login), event, wire)
	if group != "" {
		h.PublishWS(model.SubjectGroupOrders(group), event, wire)
	}
}

// PublishDeal is PublishPosition for a deal.
func (h *Handler) PublishDeal(group string, event model.EventType, d *model.Deal) {
	wire := model.NewWireDeal(d)
	h.PublishWS(model.SubjectAccountDeals(d.Login), event, wire)
	if group != "" {
		h.PublishWS(model.SubjectGroupDeals(group), event, wire)
	}
}

// entryGroup reads the account's group under the entry's lock. Callers must not hold the lock.
func entryGroup(e *book.Entry) string {
	e.Lock()
	defer e.Unlock()
	return e.Account.Group
}

// PublishTrade announces a fill in the order the trade actually happened.
//
// The position and the deals that made it go out before the order, because a client told its order
// filled before the position exists has nothing to show against it.
func (h *Handler) PublishTrade(e *book.Entry, o *model.Order, f *Fill, a *model.Account) {
	if f.Opened != nil {
		h.PublishPosition(a.Group, model.EventPositionCreate, f.Opened)
	}
	for _, p := range f.Changed {
		h.PublishPosition(a.Group, model.EventPositionUpdate, p)
	}
	for _, p := range f.Closed {
		h.PublishPosition(a.Group, model.EventPositionClose, p)
	}

	for _, d := range f.Deals {
		h.PublishDeal(a.Group, model.EventDealCreate, d)
	}

	h.PublishOrder(a.Group, model.EventOrderCreate, o)

	h.PublishAccount(a, nil)
}

// PublishAccount goes out on every tick touching the account, so it is one line rather than json.
func (h *Handler) PublishAccount(a *model.Account, positions map[int64]float64) {
	line := AccountSummary(a, positions)
	h.PublishText(model.SubjectAccountSummary(a.Login), model.EventAccountSummary, line)
	h.PublishText(model.SubjectGroupAccounts(a.Group), model.EventAccountSummary, line)
}

// SummaryFor returns "" when a tick is not worth a frame; gated per instrument so a busy symbol cannot crowd out a quiet one. Caller holds the entry's lock.
func (h *Handler) SummaryFor(e *book.Entry, symbol string, now int64, positions map[int64]float64) string {
	last, ok := e.LastSent(symbol)
	if ok && now-last.At < int64(h.summaryInterval()) {
		return ""
	}

	line := AccountSummary(e.Account, positions)
	if ok && line == last.Line {
		return ""
	}

	e.MarkSent(symbol, now, line)

	return line
}

// summaryInterval is the shortest gap between two summaries for one account on one instrument.
func (h *Handler) summaryInterval() time.Duration {
	if h.Cfg.Engine.SummaryInterval > 0 {
		return h.Cfg.Engine.SummaryInterval
	}

	return 500 * time.Millisecond
}

// PublishDealing offers one request to one dealer's queue.
func (h *Handler) PublishDealing(dealer int64, e *model.DealingEvent) {
	event := model.EventDealerRequest
	if e.EventType != model.DealingEvent_offer {
		event = model.EventDealerRequestDone
	}

	h.PublishWS(model.SubjectDealerRequests(dealer), event, e)
}
