package handler

import (
	wire "hstmodel"

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
func (h *Handler) PublishWS(subject string, event wire.EventType, payload any) {
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
func (h *Handler) PublishText(subject string, event wire.EventType, payload string) {
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

	h.PublishWS(model.SubjectAccountOrders(res.Login), wire.EventOrderRejected, res)
}

// PublishTrade announces a fill in the order the trade actually happened.
//
// The position and the deals that made it go out before the order, because a client told its order
// filled before the position exists has nothing to show against it.
func (h *Handler) PublishTrade(e *book.Entry, o *model.Order, f *Fill, a *model.Account) {
	if f.Opened != nil {
		h.PublishWS(model.SubjectAccountPositions(o.Login), wire.EventPositionCreate, f.Opened)
	}
	for _, p := range f.Changed {
		h.PublishWS(model.SubjectAccountPositions(o.Login), wire.EventPositionUpdate, p)
	}
	for _, p := range f.Closed {
		h.PublishWS(model.SubjectAccountPositions(o.Login), wire.EventPositionClose, p)
	}

	for _, d := range f.Deals {
		h.PublishWS(model.SubjectAccountDeals(d.Login), wire.EventDealCreate, d)
	}

	h.PublishWS(model.SubjectAccountOrders(o.Login), wire.EventOrderCreate, o)

	h.PublishAccount(a, nil)
}

// PublishAccount goes out on every tick touching the account, so it is one line rather than json.
func (h *Handler) PublishAccount(a *model.Account, positions map[int64]float64) {
	h.PublishText(model.SubjectAccountSummary(a.Login), wire.EventAccountSummary, AccountSummary(a, positions))
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
	event := wire.EventDealerRequest
	if e.EventType != model.DealingEvent_offer {
		event = wire.EventDealerRequestDone
	}

	h.PublishWS(model.SubjectDealerRequests(dealer), event, e)
}
