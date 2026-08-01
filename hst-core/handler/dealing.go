package handler

import (
	"context"
	"encoding/json"
	"time"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// dealingTimeout is how long a request waits for an answer before it is given up on.
// ponytail: one timeout for everyone, read it per group when settings carries the field
const dealingTimeout = 30 * time.Second

// dealingReasonMax is how much of a dealer's reason the client terminal shows.
const dealingReasonMax = 31

// Pending is one request sitting in the queue.
type Pending struct {
	Request  *model.TradeRequest
	Order    *model.Order
	Dealers  []int64
	Returned map[int64]bool
	At       int64
	Reply    string
}

// DealingSystemEventHandler takes one dealer's answer off the wire.
func (h *Handler) DealingSystemEventHandler(msg *natscore.Msg) {
	var e model.DealingEvent
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "bad dealing event", "subject", msg.Subject)
		return
	}

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.RequestId, Login: e.Login}

	switch e.EventType {
	case model.DealingEventConfirm:
		res = h.ConfirmRequest(ctx, &e)
	case model.DealingEventRequote:
		res = h.RequoteRequest(ctx, &e)
	case model.DealingEventReject:
		res = h.RejectRequest(ctx, &e)
	case model.DealingEventCancel:
		res = h.CancelRequest(ctx, &e)
	case model.DealingEventOffer:
		res = h.ReturnRequest(&e)
	default:
		res = h.refuse(res, model.RetInvalidData, "unknown dealing event")
	}

	h.reply(msg, res)
}

// SendDealing registers a request, offers it to every assigned dealer, and answers the caller queued.
func (h *Handler) SendDealing(req *model.TradeRequest, o *model.Order,
	dealers []int64) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	if len(dealers) == 0 {
		return h.refuse(res, model.RetTradeNotProcessed, "no dealer is assigned to this request")
	}

	// the row goes down first, so the request is visible in the back office while it waits
	if err := h.saveRequest(context.Background(), o); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not record a dealer request",
			"login", req.Login, "request", req.RequestId, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	p := &Pending{
		Request:  req,
		Order:    o,
		Dealers:  dealers,
		Returned: make(map[int64]bool, len(dealers)),
		At:       Now(),
	}

	h.dealingMu.Lock()
	h.dealing[req.RequestId] = p
	h.dealingMu.Unlock()

	h.offer(p, model.DealingEventOffer, 0)

	res.RetCode = int32(model.RetTradeDealerQueued)
	res.Message = model.RetTradeDealerQueued.String()
	res.OrderId = o.OrderId

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "request queued for a dealer",
		"login", req.Login, "request", req.RequestId, "dealers", len(dealers))

	return res
}

// ConfirmRequest fills the request the dealer answered. The first answer wins.
func (h *Handler) ConfirmRequest(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	p := h.takeRequest(ev.RequestId)
	if p == nil {
		return h.refuse(&model.TradeResult{RequestId: ev.RequestId, Login: ev.Login},
			model.RetNotFound, "")
	}

	res := &model.TradeResult{RequestId: p.Request.RequestId, Login: p.Request.Login}

	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	o := p.Order
	o.Dealer = ev.Dealer

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.Quotes.Get(o.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	// a queued cancel is settled by taking the order off, not by filling it
	if model.OrderState(o.State) == model.StateRequestCancel {
		e.Lock()
		h.removeOrder(ctx, e, o, "deleted [by dealer]")

		res.RetCode = int32(model.RetOK)
		res.Message = model.RetOK.String()
		res.OrderId = o.OrderId

		h.done(p, ev.Dealer)

		return res
	}

	// a pending order the dealer let through goes on the book rather than filling
	if o.Kind().Pending() {
		out := h.placeOrder(ctx, res, e, o, "dealer")
		h.done(p, ev.Dealer)

		return out
	}

	// the dealer names a price, or takes the one the client asked for, or takes the market
	price := ev.Price
	if price <= 0 {
		price = o.PriceOrder
	}
	if price <= 0 {
		price = tick.OpenPrice(o.Kind().Buy())
	}
	price = NormalisePrice(price, r.Digits)

	o.State = int32(model.StateStarted)

	e.Lock()

	// it waited on the book so the back office could see it; filling now takes it off again
	delete(e.Orders, o.OrderId)

	if code := h.checkMoney(e, o, r, tick); !code.OK() {
		e.Unlock()
		h.done(p, ev.Dealer)

		return h.refuse(res, code, "")
	}

	fill := h.Execute(e, o, r, price, Now())

	h.settle(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a confirmed request",
			"login", res.Login, "request", res.RequestId, "error", err.Error())
		h.done(p, ev.Dealer)

		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = o.OrderId
	res.Price = price
	res.Volume = o.VolumeCurrent
	res.Profit = fill.Profit

	if fill.Opened != nil {
		res.PositionId = fill.Opened.PositionId
	}
	if len(fill.Deals) > 0 {
		res.DealId = fill.Deals[0].DealId
	}

	h.done(p, ev.Dealer)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "request confirmed by a dealer",
		"login", res.Login, "request", res.RequestId, "dealer", ev.Dealer, "price", price)

	return res
}

// RequoteRequest offers the client a new price. Nothing is held: the client re-submits.
func (h *Handler) RequoteRequest(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	p := h.takeRequest(ev.RequestId)
	if p == nil {
		return h.refuse(&model.TradeResult{RequestId: ev.RequestId, Login: ev.Login},
			model.RetNotFound, "")
	}

	res := &model.TradeResult{RequestId: p.Request.RequestId, Login: p.Request.Login}

	if tick, ok := h.Quotes.Get(p.Order.Symbol); ok {
		res.Bid, res.Ask = tick.Bid, tick.Ask
	}

	// the dealer's own price replaces the side the client was asking to trade
	if ev.Price > 0 {
		if p.Order.Kind().Buy() {
			res.Ask = ev.Price
		} else {
			res.Bid = ev.Price
		}
	}

	h.dropRequest(ctx, p, "requoted")
	h.done(p, ev.Dealer)

	return h.refuse(res, model.RetTradeRequote, "")
}

// RejectRequest refuses the request with a reason the client sees.
func (h *Handler) RejectRequest(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	p := h.takeRequest(ev.RequestId)
	if p == nil {
		return h.refuse(&model.TradeResult{RequestId: ev.RequestId, Login: ev.Login},
			model.RetNotFound, "")
	}

	res := &model.TradeResult{RequestId: p.Request.RequestId, Login: p.Request.Login}

	reason := ev.Reason
	if len(reason) > dealingReasonMax {
		reason = reason[:dealingReasonMax]
	}

	h.dropRequest(ctx, p, reason)
	h.done(p, ev.Dealer)

	return h.refuse(res, model.RetTradeRejected, reason)
}

// CancelRequest removes the order the request was for.
func (h *Handler) CancelRequest(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	p := h.takeRequest(ev.RequestId)
	if p == nil {
		return h.refuse(&model.TradeResult{RequestId: ev.RequestId, Login: ev.Login},
			model.RetNotFound, "")
	}

	res := &model.TradeResult{RequestId: p.Request.RequestId, Login: p.Request.Login}

	h.dropRequest(ctx, p, "deleted [by dealer]")
	h.done(p, ev.Dealer)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = p.Order.OrderId

	return res
}

// ReturnRequest hands one dealer's copy back.
func (h *Handler) ReturnRequest(ev *model.DealingEvent) *model.TradeResult {
	res := &model.TradeResult{RequestId: ev.RequestId, Login: ev.Login}

	h.dealingMu.Lock()

	p, ok := h.dealing[ev.RequestId]
	if !ok {
		h.dealingMu.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	p.Returned[ev.Dealer] = true

	all := true
	for _, d := range p.Dealers {
		if !p.Returned[d] {
			all = false
			break
		}
	}

	if !all {
		h.dealingMu.Unlock()

		res.RetCode = int32(model.RetTradeDealerQueued)
		res.Message = model.RetTradeDealerQueued.String()

		return res
	}

	delete(h.dealing, ev.RequestId)
	h.dealingMu.Unlock()

	res.Login = p.Request.Login

	h.dropRequest(context.Background(), p, "returned by every dealer")
	h.done(p, ev.Dealer)

	return h.refuse(res, model.RetTradeDealerReturned, "")
}

// CheckDealingRequests gives up on the requests nobody answered in time.
func (h *Handler) CheckDealingRequests() {
	cutoff := Now() - int64(dealingTimeout)

	h.dealingMu.Lock()

	var stale []*Pending
	for id, p := range h.dealing {
		if p.At <= cutoff {
			stale = append(stale, p)
			delete(h.dealing, id)
		}
	}

	h.dealingMu.Unlock()

	ctx := context.Background()

	for _, p := range stale {
		h.dropRequest(ctx, p, "no dealer answered in time")
		h.done(p, 0)

		res := &model.TradeResult{RequestId: p.Request.RequestId, Login: p.Request.Login}
		h.PublishResult(h.refuse(res, model.RetTradeTimeout, ""))
	}
}

// dealersFor is who the rule hands the request to.
// ponytail: the online action takes the rule's list too, until dealer presence reaches the pod
func (h *Handler) dealersFor(d Decision, e *book.Entry) []int64 {
	return d.Dealers
}

// offer puts the request on every assigned dealer's queue.
func (h *Handler) offer(p *Pending, kind model.DealingEventType, dealer int64) {
	ev := &model.DealingEvent{
		EventType: kind,
		RequestId: p.Request.RequestId,
		Login:     p.Request.Login,
		Dealer:    dealer,
		Request:   p.Request,
		At:        Now(),
	}

	for _, d := range p.Dealers {
		h.PublishDealing(d, ev)
	}
}

// done tells every dealer the request is settled, so it leaves their queues.
func (h *Handler) done(p *Pending, dealer int64) {
	h.offer(p, model.DealingEventConfirm, dealer)
}

// takeRequest lifts a request out of the queue. Whoever gets it settles it.
func (h *Handler) takeRequest(id string) *Pending {
	h.dealingMu.Lock()
	defer h.dealingMu.Unlock()

	p, ok := h.dealing[id]
	if !ok {
		return nil
	}

	delete(h.dealing, id)

	return p
}

// dropRequest takes the order the request was for off the book.
func (h *Handler) dropRequest(ctx context.Context, p *Pending, comment string) {
	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return
	}

	e.Lock()
	h.removeOrder(ctx, e, p.Order, comment)
}

// saveRequest writes down the order a queued request is for.
func (h *Handler) saveRequest(ctx context.Context, o *model.Order) error {
	if o.OrderId > 0 {
		return h.writeOrder(ctx, o)
	}

	e, ok := h.Accounts.Get(o.Login)
	if !ok {
		return nil
	}

	e.Lock()
	account := *e.Account
	e.Unlock()

	if err := h.save(ctx, e, o, &Fill{}, &account); err != nil {
		return err
	}

	e.Lock()
	e.Orders[o.OrderId] = o
	e.Unlock()

	h.Accounts.Watch(o.Symbol, e)

	return nil
}
