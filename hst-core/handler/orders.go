package handler

import (
	"context"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// The order verbs.

// NewOrder runs one request all the way through.
func (h *Handler) NewOrder(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	r, ok := h.Settings.For(e.Account.Group, req.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.Quotes.Get(req.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	order := h.orderFrom(req, r, e.Account, tick)

	e.Lock()

	if code := h.ValidateOrder(e, order, r, tick); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	code, kind := h.checkExecution(order, r, tick)
	if !code.OK() {
		e.Unlock()
		res.Bid, res.Ask = tick.Bid, tick.Ask
		return h.refuse(res, code, "")
	}

	decision := h.Route(&Request{
		Kind:      kind,
		Order:     order,
		Entry:     e,
		Rules:     r,
		Tick:      tick,
		Gapped:    h.Quotes.Gapped(req.Symbol),
		Deviation: h.deviation(order, tick, r),
	})

	e.Unlock()

	// a rule asked to wait before deciding.
	if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}

	if !decision.Executes() {
		return h.refuseByRule(res, decision, e, req, order, model.StateRequestAdd)
	}

	// a pending order does not fill now; it goes on the book and waits for its price
	if order.Kind().Pending() {
		return h.placeOrder(ctx, res, e, order, decision.Rule.Name)
	}

	e.Lock()

	if t, ok := h.Quotes.Get(req.Symbol); ok {
		tick = t
	}
	// the standard mode checks margin when the order is placed and when a pending one triggers;
	// this group asks for one more look after the request was confirmed, because the market
	// moved while it was being decided
	if r.MarginFlags&model.MarginFlagCheckProcess != 0 {
		if code := h.checkMoney(e, order, r, tick); !code.OK() {
			e.Unlock()
			return h.refuse(res, code, "")
		}
	}

	// confirm-by-request-price fills where the client asked; confirm-by-market fills here
	price := order.PriceOrder
	if decision.AtMarket() || price <= 0 {
		price = tick.OpenPrice(order.Kind().Buy())
	}
	price = NormalisePrice(price, r.Digits)

	fill := h.Execute(e, order, r, price, Now())

	h.settle(e, order, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, order, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a trade",
			"login", req.Login, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = order.OrderId
	res.Price = price
	res.Volume = order.VolumeCurrent
	res.Profit = fill.Profit
	res.Rule = decision.Rule.Name

	if fill.Opened != nil {
		res.PositionId = fill.Opened.PositionId
	}
	if len(fill.Deals) > 0 {
		res.DealId = fill.Deals[0].DealId
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "trade done",
		"login", req.Login, "symbol", req.Symbol, "volume", model.Lots(order.VolumeCurrent),
		"price", price, "rule", decision.Rule.Name)

	return res
}

// UpdateOrder moves a working pending order's price, levels and expiry.
func (h *Handler) UpdateOrder(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

	o, ok := e.Orders[req.OrderId]
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}
	if !model.OrderState(o.State).Live() {
		e.Unlock()
		return h.refuse(res, model.RetTradeFrozen, "")
	}
	if o.Kind().Market() {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "a market order cannot be modified")
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.Quotes.Get(o.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	// the change is validated and routed on a copy, so a refusal leaves the working order alone
	want := *o
	if req.Price > 0 {
		want.PriceOrder = req.Price
	}
	if req.PriceTrigger > 0 {
		want.PriceTrigger = req.PriceTrigger
	}
	want.PriceSL, want.PriceTP = req.PriceSL, req.PriceTP
	if req.TypeTime > 0 || req.Expiry > 0 {
		want.TypeTime, want.TimeExpiration = req.TypeTime, req.Expiry
	}

	if code := h.checkExpiry(&want, r); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}
	if code := h.checkStops(&want, r, tick); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	decision := h.Route(&Request{
		Kind:      model.RouteModify,
		Order:     &want,
		Entry:     e,
		Rules:     r,
		Tick:      tick,
		Gapped:    h.Quotes.Gapped(o.Symbol),
		Deviation: h.deviation(&want, tick, r),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, &want, model.StateRequestModify)
	}

	*o = want
	if req.Dealer != 0 {
		o.Dealer = req.Dealer
	}

	saved := *o
	e.Unlock()

	if err := h.writeOrder(ctx, &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not modify an order",
			"login", saved.Login, "order", saved.OrderId, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	h.PublishWS(model.SubjectAccountOrders(saved.Login), "order", &saved)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = saved.OrderId
	res.Price = saved.PriceOrder
	res.Volume = saved.VolumeCurrent
	res.Rule = decision.Rule.Name

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "order modified",
		"login", saved.Login, "order", saved.OrderId, "price", saved.PriceOrder)

	return res
}

// CancelOrder is the client's own removal of a working pending order.
func (h *Handler) CancelOrder(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

	o, ok := e.Orders[req.OrderId]
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}
	if !model.OrderState(o.State).Live() {
		e.Unlock()
		return h.refuse(res, model.RetTradeFrozen, "")
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, _ := h.Quotes.Get(o.Symbol)

	decision := h.Route(&Request{
		Kind: model.RouteRemove, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(o.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.StateRequestCancel)
	}

	comment := req.Comment
	if comment == "" {
		comment = "deleted [by client]"
	}

	orderId, volume := o.OrderId, o.VolumeCurrent
	h.removeOrder(ctx, e, o, comment)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = orderId
	res.Volume = volume
	res.Rule = decision.Rule.Name

	return res
}

// CookOrder fills a pending order the price reached, or turns a stop limit into a limit.
func (h *Handler) CookOrder(ctx context.Context, e *book.Entry, hit pendingHit, t model.Tick) error {
	o := hit.order

	e.Lock()

	// another tick may have taken it already
	if _, still := e.Orders[o.OrderId]; !still {
		e.Unlock()
		return nil
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return nil
	}

	// expiry first: an order whose time has run out must not fill on the tick that expires it
	if o.ActivationFlags&model.ActivationFlagNoExpiry == 0 &&
		o.TimeExpiration > 0 && o.TimeExpiration <= Now() {
		h.removeOrder(ctx, e, o, "expired")
		return nil
	}

	if hit.toLimit {
		h.cookStopLimit(ctx, e, o, r)
		return nil
	}

	// the rules see an activation as its own kind of request, so a broker can hold one during a gap
	decision := h.Route(&Request{
		Kind: model.RouteActivate, Order: o, Entry: e, Rules: r, Tick: t,
		Gapped: h.Quotes.Gapped(o.Symbol),
	})

	if !decision.Executes() {
		// cancel order is the rule that exists precisely for this.
		if decision.Action == model.ActionCancelOrder {
			h.removeOrder(ctx, e, o, "deleted [by routing rule]")
			return nil
		}

		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "an activation was not admitted by any rule",
			"login", o.Login, "order", o.OrderId)

		return nil
	}

	// the margin is checked again when the order fires, not only when it was accepted
	if code := h.checkMoney(e, o, r, t); !code.OK() {
		h.removeOrder(ctx, e, o, "canceled, not enough money")
		return nil
	}

	// a limit fills at its own price; anything else fills at the market
	price := o.PriceOrder
	if decision.AtMarket() && o.Kind() != model.OrderBuyLimit && o.Kind() != model.OrderSellLimit {
		price = t.OpenPrice(o.Kind().Buy())
	}
	price = NormalisePrice(price, r.Digits)

	// the fill is a market order of the same side and size, against the pending order's ticket
	fillOrder := *o
	fillOrder.Type = int32(model.OrderBuy)
	if !o.Kind().Buy() {
		fillOrder.Type = int32(model.OrderSell)
	}
	fillOrder.Reason = int32(model.ReasonClient)

	fill := h.Execute(e, &fillOrder, r, price, Now())

	delete(e.Orders, o.OrderId)
	o.State = int32(model.StateFilled)
	o.TimeDone = Now()
	o.PriceCurrent = price

	h.settle(e, &fillOrder, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save an activation",
			"login", o.Login, "order", o.OrderId, "error", err.Error())
		return err
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "pending order filled",
		"login", o.Login, "order", o.OrderId, "type", model.OrderType(o.Type),
		"price", price)

	return nil
}

// SaveOrderAndPublish writes the order and everything the fill produced, then announces it.
func (h *Handler) SaveOrderAndPublish(ctx context.Context, e *book.Entry, o *model.Order,
	f *Fill, a *model.Account) error {
	if o.OrderId > 0 {
		if err := h.writeOrder(ctx, o); err != nil {
			return err
		}
		if err := h.saveFill(ctx, e, o, f, a); err != nil {
			return err
		}
	} else if err := h.save(ctx, e, o, f, a); err != nil {
		return err
	}

	h.PublishTrade(e, o, f, a)

	return nil
}

// SaveOrderAndPublishAsync hands the write to the worker pool.
func (h *Handler) SaveOrderAndPublishAsync(e *book.Entry, o *model.Order, f *Fill, a *model.Account) {
	h.Workers.Submit(func(ctx context.Context) {
		if err := h.SaveOrderAndPublish(ctx, e, o, f, a); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a trade",
				"login", o.Login, "order", o.OrderId, "error", err.Error())
		}
	})
}

// placeOrder puts a working order on the book and leaves it there.
func (h *Handler) placeOrder(ctx context.Context, res *model.TradeResult, e *book.Entry,
	o *model.Order, rule string) *model.TradeResult {
	o.State = int32(model.StatePlaced)

	e.Lock()
	account := *e.Account
	e.Unlock()

	if err := h.save(ctx, e, o, &Fill{}, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not place an order",
			"login", o.Login, "error", err.Error())
		return h.refuse(res, model.RetError, "")
	}

	e.Lock()
	e.Orders[o.OrderId] = o
	e.Unlock()

	h.Accounts.Watch(o.Symbol, e)
	h.PublishWS(model.SubjectAccountOrders(o.Login), "order", o)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = o.OrderId
	res.Price = o.PriceOrder
	res.Volume = o.VolumeCurrent
	res.Rule = rule

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "order placed",
		"login", o.Login, "order", o.OrderId, "type", o.Kind(), "price", o.PriceOrder)

	return res
}

// writeOrder puts a changed order back, in its own transaction.
func (h *Handler) writeOrder(ctx context.Context, o *model.Order) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := updateOrder(ctx, tx, o); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// settle puts the fill onto the account and recomputes the money.
func (h *Handler) settle(e *book.Entry, o *model.Order, f *Fill, r *settings.Rules) {
	// commission comes off as the deal is booked, so the client sees the true cost of the trade
	for _, d := range f.Deals {
		d.Commission = -h.CommissionFor(d, r)
		e.Account.Balance += d.Commission
	}

	// Only the swap comes off a closed position here.
	for _, p := range f.Closed {
		delete(e.Positions, p.PositionId)
		e.Account.Balance += p.Storage
	}

	h.creditRealised(e, f.Profit)

	if f.Opened != nil {
		// the id is filled in when the row is written.
		f.Opened.PositionId = h.nextTempId()
		e.Positions[f.Opened.PositionId] = f.Opened
	}

	o.State = int32(model.StateFilled)
	o.TimeDone = Now()
	o.PriceCurrent = f.Price

	h.SettleAccount(e).Apply(e.Account)
}

// orderFrom builds the order record a request is asking for.
func (h *Handler) orderFrom(req *model.TradeRequest, r *settings.Rules, a *model.Account, t model.Tick) *model.Order {
	now := Now()

	return &model.Order{
		Login:          req.Login,
		Dealer:         req.Dealer,
		Symbol:         req.Symbol,
		Digits:         r.Digits,
		ContractSize:   r.ContractSize,
		State:          int32(model.StateStarted),
		Reason:         req.Reason,
		TimeSetup:      now,
		TimeExpiration: req.Expiry,
		Type:           req.Type,
		TypeFill:       req.TypeFill,
		TypeTime:       req.TypeTime,
		PriceOrder:     req.Price,
		PriceTrigger:   req.PriceTrigger,
		PriceCurrent:   t.OpenPrice(model.OrderType(req.Type).Buy()),
		PriceSL:        req.PriceSL,
		PriceTP:        req.PriceTP,
		VolumeInitial:  req.Volume,
		VolumeCurrent:  req.Volume,
		ExpertId:       req.ExpertId,
		PositionId:     req.PositionId,
		PositionById:   req.PositionById,
		Comment:        req.Comment,
		RateMargin:     h.RateMargin(r, a, model.OrderType(req.Type).Buy()),
	}
}

// kindOf is which routing request type this order counts as, which decides what rules see it.
func (h *Handler) kindOf(o *model.Order, r *settings.Rules) model.RouteFlags {
	if o.Kind().Pending() {
		return model.RoutePending
	}
	if o.Kind() == model.OrderCloseBy {
		return model.RouteCloseBy
	}

	switch r.ExecMode {
	case model.ExecInstant:
		return model.RouteInstant
	case model.ExecRequest:
		return model.RouteRequest
	case model.ExecExchange:
		return model.RouteExchange
	default:
		return model.RouteMarket
	}
}

// deviation is how far the requested price is from the market, in points.
func (h *Handler) deviation(o *model.Order, t model.Tick, r *settings.Rules) float64 {
	if o.PriceOrder <= 0 || r.Point <= 0 {
		return 0
	}

	if o.Kind().Buy() {
		return Points(t.Ask-o.PriceOrder, r.Point)
	}

	return Points(o.PriceOrder-t.Bid, r.Point)
}

// refuse fills in a refusal with its return code.
func (h *Handler) refuse(res *model.TradeResult, code model.RetCode, message string) *model.TradeResult {
	res.RetCode = int32(code)
	res.Message = message
	if res.Message == "" {
		res.Message = code.String()
	}

	h.Log.Log(logger.TypeTrade, logger.CodeWarn, "trade refused",
		"login", res.Login, "request", res.RequestId,
		"retcode", int32(code), "reason", code.String(), "rule", res.Rule)

	return res
}

// refuseByRule turns a routing decision that did not execute into a refusal, or hands it to a dealer.
func (h *Handler) refuseByRule(res *model.TradeResult, d Decision, e *book.Entry,
	req *model.TradeRequest, o *model.Order, state model.OrderState) *model.TradeResult {
	if d.Rule == nil {
		return h.refuse(res, model.RetTradeNotProcessed, "")
	}

	res.Rule = d.Rule.Name

	switch d.Action {
	case model.ActionReject:
		return h.refuse(res, model.RetTradeRejected, d.Reason)
	case model.ActionRequote:
		return h.refuse(res, model.RetTradeRequote, "")
	case model.ActionDealer, model.ActionDealerOnline:
		o.State = int32(state)
		return h.SendDealing(req, o, d.Dealers, d.Rule.RoutingId)
	case model.ActionCancelOrder:
		return h.refuse(res, model.RetTradeRejected, "order cancelled")
	}

	return h.refuse(res, model.RetTradeNotProcessed, "")
}

// nextTempId hands out the temporary keys a position is held under until its row is written.
func (h *Handler) nextTempId() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.tempId--

	return h.tempId
}

// creditRealised puts a realised result where the group says it belongs.
//
// In the day profit and loss mode a profit is held aside until the end of the day rather than
// being available to trade on straight away; a loss always lands at once.
func (h *Handler) creditRealised(e *book.Entry, amount float64) {
	g, ok := h.Settings.Group(e.Account.Group)

	if amount > 0 && ok && g.MarginFreeProfit == int32(model.FreeMarginDayProfitLoss) {
		e.Account.BlockedProfit += amount
		return
	}

	e.Account.Balance += amount
}
