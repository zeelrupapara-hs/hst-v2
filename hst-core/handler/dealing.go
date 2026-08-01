package handler

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"hstcore/internal/book"

	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// dealingTimeout is how long a request waits when the instrument names no timeout of its own.
const dealingTimeout = 30 * time.Second

// dealingSweepEvery is how often the queue is checked for requests that ran out of time.
const dealingSweepEvery = time.Second

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

	// Requoted is the price a dealer offered back, while it waits for the client to take it.
	Requoted float64

	// Holder is where in Dealers the request currently sits. The list is the rule's own order,
	// so the request travels down it: whoever holds it answers, and passing it back moves it on.
	Holder int
}

// holder is the dealer the request is with, or 0 when it has run off the end of the list.
func (p *Pending) holder() int64 {
	if p.Holder < 0 || p.Holder >= len(p.Dealers) {
		return 0
	}

	return p.Dealers[p.Holder]
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
	case model.DealingEventAccept:
		res = h.AcceptRequote(ctx, &e)
	case model.DealingEventConfirm:
		res = h.ConfirmRequest(ctx, &e)
	case model.DealingEventRequote:
		res = h.RequoteRequest(ctx, &e)
	case model.DealingEventReject:
		res = h.RejectRequest(ctx, &e)
	case model.DealingEventCancel:
		res = h.CancelRequest(ctx, &e)
	case model.DealingEventReturn, model.DealingEventOffer:
		res = h.ReturnRequest(&e)
	default:
		res = h.refuse(res, model.RetInvalidData, "unknown dealing event")
	}

	h.reply(msg, res)
}

// SendDealing registers a request, offers it to every assigned dealer, and answers the caller queued.
func (h *Handler) SendDealing(req *model.TradeRequest, o *model.Order,
	dealers []int64, routingId int64) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	if len(dealers) == 0 {
		return h.refuse(res, model.RetTradeNotProcessed, "no dealer is assigned to this request")
	}

	o.State = int32(model.StateRequestAdd)
	if o.OrderId > 0 {
		o.State = int32(model.StateRequestModify)
	}

	// the rule is written down so the desk can show each dealer only their own queue
	o.RoutingId = routingId

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

	// a dealer can sit on a request for a while, so the group can ask for margin to be looked
	// at once more before their answer is acted on
	if r.MarginFlags&model.MarginFlagCheckProcess != 0 {
		if code := h.checkMoney(e, o, r, tick); !code.OK() {
			e.Unlock()
			h.done(p, ev.Dealer)

			return h.refuse(res, code, "")
		}
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

// RequoteRequest offers the client a new price and holds the request open at it, so the client
// can take it without placing the order again.
func (h *Handler) RequoteRequest(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	h.dealingMu.Lock()

	p, ok := h.dealing[ev.RequestId]
	if !ok {
		h.dealingMu.Unlock()
		return h.refuse(&model.TradeResult{RequestId: ev.RequestId, Login: ev.Login},
			model.RetNotFound, "")
	}

	// the clock restarts: the price the client is being shown is this dealer's, from now
	p.Requoted = ev.Price
	p.At = Now()

	h.dealingMu.Unlock()

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

	return h.refuse(res, model.RetTradeRequote, "")
}

// AcceptRequote is the client taking a price a dealer offered back.
//
// Whether that ends the matter depends on the instrument. In request execution the dealer
// confirms once more if the group asked for it. In instant execution the order goes back to the
// dealer as well, unless the group turned on fast confirmation and the price the client is
// taking sits inside the deviation they themselves allowed.
func (h *Handler) AcceptRequote(ctx context.Context, ev *model.DealingEvent) *model.TradeResult {
	h.dealingMu.Lock()
	p, ok := h.dealing[ev.RequestId]
	h.dealingMu.Unlock()

	res := &model.TradeResult{RequestId: ev.RequestId, Login: ev.Login}

	if !ok || p.Requoted <= 0 {
		return h.refuse(res, model.RetNotFound, "")
	}

	res.Login = p.Request.Login

	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	r, ok := h.Settings.For(e.Account.Group, p.Order.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	if h.needsSecondConfirmation(p, r) {
		h.dealingMu.Lock()
		p.At = Now()
		h.dealingMu.Unlock()

		// the price the client took is the one the dealer is being asked to stand behind
		p.Order.PriceOrder = p.Requoted
		h.offer(p, model.DealingEventOffer, 0)

		h.Log.Log(logger.TypeTrade, logger.CodeOK, "requote accepted, back to the dealer",
			"login", res.Login, "request", res.RequestId, "price", p.Requoted)

		return h.refuse(res, model.RetTradeDealerQueued, "")
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "requote accepted, filled without the dealer",
		"login", res.Login, "request", res.RequestId, "price", p.Requoted)

	return h.ConfirmRequest(ctx, &model.DealingEvent{
		EventType: model.DealingEventConfirm,
		RequestId: ev.RequestId,
		Login:     p.Request.Login,
		Dealer:    p.Order.Dealer,
		Price:     p.Requoted,
	})
}

// needsSecondConfirmation decides whether an accepted requote goes back to the dealer.
func (h *Handler) needsSecondConfirmation(p *Pending, r *settings.Rules) bool {
	if r.ExecMode == model.ExecInstant {
		if r.IEFlags&model.InstantFlagFastConfirmation == 0 {
			return true
		}

		// fast confirmation only covers a price the client had already said they would take
		return !h.withinClientDeviation(p, r)
	}

	return r.REFlags&model.RequestFlagOrder != 0
}

// withinClientDeviation reports whether the requoted price is inside the slippage the client
// allowed when they sent the order.
func (h *Handler) withinClientDeviation(p *Pending, r *settings.Rules) bool {
	if p.Request.Deviation <= 0 || r.Point <= 0 {
		return false
	}

	asked := p.Request.Price
	if asked <= 0 {
		return false
	}

	return Points(math.Abs(p.Requoted-asked), r.Point) <= float64(p.Request.Deviation)
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

	// only whoever is holding it can hand it back
	if p.holder() != ev.Dealer {
		h.dealingMu.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	p.Returned[ev.Dealer] = true

	// down the list to the next dealer who can take it
	next := h.nextHolder(p)
	if next >= 0 {
		p.Holder = next
		p.At = Now()
		h.dealingMu.Unlock()

		h.offer(p, model.DealingEventOffer, 0)

		h.Log.Log(logger.TypeTrade, logger.CodeOK, "request passed to the next dealer",
			"login", p.Request.Login, "request", ev.RequestId,
			"from", ev.Dealer, "to", p.holder())

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

// nextHolder is the position of the next dealer below the current one who has not already
// handed the request back. It returns -1 when the list is exhausted.
//
// Presence is not re-checked here: the rule's action already decided whether the list was
// narrowed to dealers who had connected, and re-reading it would quietly change what the
// broker configured.
//
// Called with the dealing lock held.
func (h *Handler) nextHolder(p *Pending) int {
	for i := p.Holder + 1; i < len(p.Dealers); i++ {
		if !p.Returned[p.Dealers[i]] {
			return i
		}
	}

	return -1
}

// runDealingSweep gives up on unanswered requests, on a tick short enough that the shortest
// instrument timeout is still roughly honoured.
func (h *Handler) runDealingSweep(ctx context.Context) {
	t := time.NewTicker(dealingSweepEvery)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.CheckDealingRequests()
		}
	}
}

// CheckDealingRequests gives up on the requests nobody answered in time.
func (h *Handler) CheckDealingRequests() {
	now := Now()

	h.dealingMu.Lock()

	var stale []*Pending
	for id, p := range h.dealing {
		if now-p.At >= int64(h.dealingTimeout(p)) {
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

// dealingTimeout is how long the dealer's price stays good for this instrument, falling back to
// the engine default when the group set none.
func (h *Handler) dealingTimeout(p *Pending) time.Duration {
	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return dealingTimeout
	}

	r, ok := h.Settings.For(e.Account.Group, p.Request.Symbol)
	if !ok || r.RequestTimeout <= 0 {
		return dealingTimeout
	}

	return time.Duration(r.RequestTimeout) * time.Second
}

// offer puts the request on the queue of the dealer whose turn it is.
func (h *Handler) offer(p *Pending, kind model.DealingEventType, dealer int64) {
	ev := &model.DealingEvent{
		EventType: kind,
		RequestId: p.Request.RequestId,
		Login:     p.Request.Login,
		Dealer:    dealer,
		Request:   p.Request,
		At:        Now(),
	}

	if to := p.holder(); to != 0 {
		h.PublishDealing(to, ev)
	}
}

// done tells everyone who has seen the request that it is settled, so it leaves their queues.
func (h *Handler) announceDone(p *Pending, dealer int64) {
	ev := &model.DealingEvent{
		EventType: model.DealingEventConfirm,
		RequestId: p.Request.RequestId,
		Login:     p.Request.Login,
		Dealer:    dealer,
		Request:   p.Request,
		At:        Now(),
	}

	// everyone up to and including the holder has had it on their screen at some point
	for i := 0; i <= p.Holder && i < len(p.Dealers); i++ {
		h.PublishDealing(p.Dealers[i], ev)
	}
}

func (h *Handler) done(p *Pending, dealer int64) {
	h.announceDone(p, dealer)
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

// RecoverDealingRequests puts the requests that were waiting on a dealer back on the desk after
// a restart.
//
// The order row survives a pod going down; the queue it was sitting in does not. Without this
// the request would stay in its request state for good, answerable by nobody and swept by
// nothing. The rule that queued it was written down with it, so it goes back to the same desk
// with its clock started again.
func (h *Handler) RecoverDealingRequests() {
	var back, dropped int

	h.Accounts.Each(func(e *book.Entry) {
		e.Lock()
		waiting := make([]*model.Order, 0, 2)
		for _, o := range e.Orders {
			if model.OrderState(o.State).AwaitingDealer() {
				waiting = append(waiting, o)
			}
		}
		e.Unlock()

		for _, o := range waiting {
			dealers := h.dealersOfRule(o.RoutingId)
			if len(dealers) == 0 {
				// the rule is gone, or names nobody: nothing can answer this any more.
				// removeOrder takes the entry locked and gives it back unlocked.
				e.Lock()
				h.removeOrder(context.Background(), e, o, "no dealer is assigned")
				dropped++
				continue
			}

			p := &Pending{
				Request: &model.TradeRequest{
					RequestId: recoveredRequestId(o.OrderId),
					Login:     o.Login,
					Symbol:    o.Symbol,
					Type:      o.Type,
					Volume:    o.VolumeCurrent,
					Price:     o.PriceOrder,
					PriceSL:   o.PriceSL,
					PriceTP:   o.PriceTP,
					TypeFill:  o.TypeFill,
					TypeTime:  o.TypeTime,
					Expiry:    o.TimeExpiration,
					Comment:   o.Comment,
					ExpertId:  o.ExpertId,
					Reason:    o.Reason,
				},
				Order:    o,
				Dealers:  dealers,
				Returned: make(map[int64]bool, len(dealers)),
				At:       Now(),
			}

			h.dealingMu.Lock()
			h.dealing[p.Request.RequestId] = p
			h.dealingMu.Unlock()

			h.offer(p, model.DealingEventOffer, 0)
			back++
		}
	})

	if back > 0 || dropped > 0 {
		h.Log.Log(logger.TypeTrade, logger.CodeOK, "dealer requests recovered",
			"requeued", back, "dropped", dropped)
	}
}

// dealersOfRule is the desk a rule still names.
func (h *Handler) dealersOfRule(routingId int64) []int64 {
	if routingId == 0 {
		return nil
	}

	for i := range h.rules {
		if h.rules[i].RoutingId == routingId {
			return h.rules[i].Dealers
		}
	}

	return nil
}

// recoveredRequestId names a request whose original id went down with the pod.
func recoveredRequestId(orderId int64) string {
	return "recovered." + strconv.FormatInt(orderId, 10)
}
