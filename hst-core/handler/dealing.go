package handler

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/shardmap"

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
	Request *model.TradeRequest
	Order   *model.Order
	// Target is the working order a modification is queued against; Order then holds what it should become.
	Target   *model.Order
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

// live is the order sitting on the book for this request: the target of a modification, or the order itself.
func (p *Pending) live() *model.Order {
	if p.Target != nil {
		return p.Target
	}
	return p.Order
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

	if h.forwarded(msg, "dealing", e.Login) {
		return
	}

	ctx := context.Background()
	res := &model.TradeResult{RequestId: e.RequestId, Login: e.Login}

	// a command naming a login acts on that login's request only
	if e.Login != 0 {
		h.dealingMu.Lock()
		p, ok := h.dealing[h.dealingKey(e.RequestId)]
		h.dealingMu.Unlock()
		if ok && p.Request.Login != e.Login {
			h.reply(msg, h.refuse(res, model.RetTradeAccountNotFound, "the request belongs to another account"))
			return
		}
	}

	switch e.EventType {
	case model.DealingEvent_accept:
		res = h.AcceptRequote(ctx, &e)
	case model.DealingEvent_confirm:
		res = h.ConfirmRequest(ctx, &e)
	case model.DealingEvent_requote:
		res = h.RequoteRequest(ctx, &e)
	case model.DealingEvent_reject:
		res = h.RejectRequest(ctx, &e)
	case model.DealingEvent_cancel:
		res = h.CancelRequest(ctx, &e)
	case model.DealingEvent_return, model.DealingEvent_offer:
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

	// the caller says what kind of request this is; a fresh order is an add
	if !o.State.IsAwaitingDealer() {
		o.State = model.OrderState_request_add
		if o.OrderId > 0 {
			o.State = model.OrderState_request_modify
		}
	}

	// the rule is written down so the desk can show each dealer only their own queue
	o.RoutingId = routingId

	p := &Pending{
		Request:  req,
		Order:    o,
		Dealers:  dealers,
		Returned: make(map[int64]bool, len(dealers)),
		At:       Now(),
	}

	// a modification is queued against the working order, which keeps its own prices until a dealer confirms
	if o.State == model.OrderState_request_modify && o.OrderId > 0 {
		if e, ok := h.Accounts.Get(o.Login); ok {
			e.Lock()
			if live := e.Orders[o.OrderId]; live != nil && live != o {
				live.State = model.OrderState_request_modify
				live.RoutingId = routingId
				p.Target = live
			}
			e.Unlock()
		}
	}

	// the row goes down first, so the request is visible in the back office while it waits
	if err := h.saveRequest(context.Background(), p.live()); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not record a dealer request",
			"login", req.Login, "request", req.RequestId, "error", err.Error())
		if e, ok := h.Accounts.Get(o.Login); ok && p.Target != nil {
			e.Lock()
			p.Target.State = model.OrderState_placed
			e.Unlock()
		}
		return h.refuse(res, model.RetError, "")
	}

	h.dealingMu.Lock()
	h.dealing[req.RequestId] = p
	h.dealingMu.Unlock()

	h.offer(p, model.DealingEvent_offer, 0)

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
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	o := p.Order
	o.Dealer = ev.Dealer

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.QuoteFor(r, o.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	// a queued cancel is settled by taking the order off, not by filling it
	if o.State == model.OrderState_request_cancel {
		if !h.lockHeld(e) {
			return h.refuse(res, model.RetTradeAccountNotFound, "")
		}
		h.removeOrder(ctx, e, o, "deleted [by dealer]")

		res.RetCode = int32(model.RetOK)
		res.Message = model.RetOK.String()
		res.OrderId = o.OrderId

		h.done(p, ev.Dealer)

		return res
	}

	// a queued level change is written to the position, never filled
	if o.PositionId != 0 && o.VolumeCurrent == 0 {
		return h.confirmLevels(ctx, res, e, p, ev.Dealer)
	}

	// a queued modification is applied to the working order now, and only now
	if p.Target != nil {
		if !h.lockHeld(e) {
			return h.refuse(res, model.RetTradeAccountNotFound, "")
		}
		before := *p.Target
		want := *o
		want.State = model.OrderState_placed
		*p.Target = want
		saved := *p.Target
		e.Unlock()

		if err := h.writeOrder(ctx, &saved); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not modify an order",
				"login", saved.Login, "order", saved.OrderId, "error", err.Error())
			e.Lock()
			*p.Target = before
			e.Unlock()
			h.done(p, ev.Dealer)
			return h.refuse(res, model.RetError, "")
		}

		h.PublishOrder(entryGroup(e), model.EventOrderUpdate, &saved)

		res.RetCode = int32(model.RetOK)
		res.Message = model.RetOK.String()
		res.OrderId = saved.OrderId
		res.Price = saved.PriceOrder
		res.Volume = saved.VolumeCurrent

		h.done(p, ev.Dealer)

		return res
	}

	// a pending order the dealer let through goes on the book rather than filling
	if o.Kind().IsPending() {
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
		price = tick.OpenPrice(o.Kind().IsBuy())
	}
	price = NormalisePrice(price, r.Digits)

	o.State = model.OrderState_started

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	snapshot := e.Snapshot()

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

	h.bookFill(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a confirmed request",
			"login", res.Login, "request", res.RequestId, "error", err.Error())
		h.restore(e, snapshot)
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

	p, ok := h.dealing[h.dealingKey(ev.RequestId)]
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

	// the requote shows the client the price their own group would be charged
	if e, ok := h.Accounts.Get(p.Request.Login); ok {
		if r, ok := h.Settings.For(e.Account.Group, p.Order.Symbol); ok {
			if tick, ok := h.QuoteFor(r, p.Order.Symbol); ok {
				res.Bid, res.Ask = tick.Bid, tick.Ask
			}
		}
	}

	// the dealer's own price replaces the side the client was asking to trade
	if ev.Price > 0 {
		if p.Order.Kind().IsBuy() {
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
	p, ok := h.dealing[h.dealingKey(ev.RequestId)]
	h.dealingMu.Unlock()

	res := &model.TradeResult{RequestId: ev.RequestId, Login: ev.Login}

	if !ok || p.Requoted <= 0 {
		return h.refuse(res, model.RetNotFound, "")
	}

	res.Login = p.Request.Login

	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
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
		h.offer(p, model.DealingEvent_offer, 0)

		h.Log.Log(logger.TypeTrade, logger.CodeOK, "requote accepted, back to the dealer",
			"login", res.Login, "request", res.RequestId, "price", p.Requoted)

		return h.refuse(res, model.RetTradeDealerQueued, "")
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "requote accepted, filled without the dealer",
		"login", res.Login, "request", res.RequestId, "price", p.Requoted)

	return h.ConfirmRequest(ctx, &model.DealingEvent{
		EventType: model.DealingEvent_confirm,
		RequestId: ev.RequestId,
		Login:     p.Request.Login,
		Dealer:    p.Order.Dealer,
		Price:     p.Requoted,
	})
}

// needsSecondConfirmation decides whether an accepted requote goes back to the dealer.
func (h *Handler) needsSecondConfirmation(p *Pending, r *settings.Rules) bool {
	if r.ExecMode == model.ExecMode_instant {
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

	p, ok := h.dealing[h.dealingKey(ev.RequestId)]
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

		h.offer(p, model.DealingEvent_offer, 0)

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

// RunDealingSweep gives up on unanswered requests, on a tick short enough that the shortest
// instrument timeout is still roughly honoured.
func (h *Handler) RunDealingSweep(ctx context.Context) {
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
		h.PublishRejected(h.refuse(res, model.RetTradeTimeout, ""))
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
		EventType: model.DealingEvent_confirm,
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

// dealingKey resolves the queue key: the engine's request id, or the order id a client can see.
// Called with the dealing lock held.
func (h *Handler) dealingKey(id string) string {
	if _, ok := h.dealing[id]; ok {
		return id
	}

	orderId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return id
	}

	for key, p := range h.dealing {
		if p.Order != nil && p.Order.OrderId == orderId {
			return key
		}
	}

	return id
}

func (h *Handler) takeRequest(id string) *Pending {
	h.dealingMu.Lock()
	defer h.dealingMu.Unlock()

	id = h.dealingKey(id)

	p, ok := h.dealing[id]
	if !ok {
		return nil
	}

	delete(h.dealing, id)

	return p
}

// dropRequest settles a request nobody confirmed: a new order comes off the book, a queued change
// or cancel leaves the working order exactly as it was.
func (h *Handler) dropRequest(ctx context.Context, p *Pending, comment string) {
	e, ok := h.Accounts.Get(p.Request.Login)
	if !ok {
		return
	}

	o := p.live()

	if !h.lockHeld(e) {
		return
	}

	if o.State == model.OrderState_request_add {
		h.removeOrder(ctx, e, o, comment)
		return
	}

	o.State = model.OrderState_placed
	saved := *o
	e.Unlock()

	if err := h.writeOrder(ctx, &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not put an order back to working",
			"login", saved.Login, "order", saved.OrderId, "error", err.Error())
		return
	}

	h.PublishOrder(entryGroup(e), model.EventOrderUpdate, &saved)
}

// dropDealingFor forgets the requests of accounts in the shards given, so the pod taking them on can re-queue them.
func (h *Handler) dropDealingFor(shards map[uint32]bool) {
	h.dealingMu.Lock()
	var moved []*Pending
	for id, p := range h.dealing {
		if shards[shardmap.ShardOf(p.Request.Login)] {
			moved = append(moved, p)
			delete(h.dealing, id)
		}
	}
	h.dealingMu.Unlock()

	// off the dealers' screens here; the new owner offers it again under its own id
	for _, p := range moved {
		h.announceDone(p, 0)
	}
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
// a restart, or for the shards just gained when the set is not nil.
//
// The order row survives a pod going down; the queue it was sitting in does not. Without this
// the request would stay in its request state for good, answerable by nobody and swept by
// nothing. The rule that queued it was written down with it, so it goes back to the same desk
// with its clock started again.
func (h *Handler) RecoverDealingRequests(shards map[uint32]bool) {
	var back, dropped int

	h.Accounts.Each(func(e *book.Entry) {
		if shards != nil && !shards[shardmap.ShardOf(e.Account.Login)] {
			return
		}

		e.Lock()
		waiting := make([]*model.Order, 0, 2)
		for _, o := range e.Orders {
			if o.State.IsAwaitingDealer() {
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

			// a queued change has no record of what was asked, so the order is simply working again
			if o.State != model.OrderState_request_add {
				e.Lock()
				o.State = model.OrderState_placed
				saved := *o
				e.Unlock()
				if err := h.writeOrder(context.Background(), &saved); err != nil {
					h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not put an order back to working",
						"login", saved.Login, "order", saved.OrderId, "error", err.Error())
				}
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
					ExpiryAt:  o.TimeExpiration,
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

			h.offer(p, model.DealingEvent_offer, 0)
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

	h.mu.RLock()
	rules := h.rules
	h.mu.RUnlock()

	for i := range rules {
		if rules[i].RoutingId == routingId {
			return rules[i].Dealers
		}
	}

	return nil
}

// recoveredRequestId names a request whose original id went down with the pod.
func recoveredRequestId(orderId int64) string {
	return "recovered." + strconv.FormatInt(orderId, 10)
}

// confirmLevels applies a dealer-approved SL/TP change to the position and takes the request row off the book.
func (h *Handler) confirmLevels(ctx context.Context, res *model.TradeResult, e *book.Entry, p *Pending, dealer int64) *model.TradeResult {
	o := p.Order

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	pos := h.positionById(e, o.PositionId)
	if pos == nil {
		h.removeOrder(ctx, e, o, "position gone [by dealer]")
		h.done(p, dealer)
		return h.refuse(res, model.RetNotFound, "")
	}

	before := *pos
	pos.PriceSL, pos.PriceTP = o.PriceSL, o.PriceTP
	pos.TimeUpdate = Now()
	saved := *pos
	e.Unlock()

	if err := h.SavePositionAndPublish(ctx, entryGroup(e), &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not modify a position",
			"login", saved.Login, "position", saved.PositionId, "error", err.Error())
		e.Lock()
		*pos = before
		e.Unlock()
		h.done(p, dealer)
		return h.refuse(res, model.RetError, "")
	}

	e.Lock()
	h.removeOrder(ctx, e, o, "levels set [by dealer]")

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.PositionId = saved.PositionId
	res.Volume = saved.Volume

	h.done(p, dealer)

	return res
}
