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
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	r, ok := h.Settings.For(e.Account.Group, req.Symbol)
	if !ok {
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.QuoteFor(r, req.Symbol)
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
		return h.refuseByRule(res, decision, e, req, order, model.OrderState_request_add)
	}

	// a pending order does not fill now; it goes on the book and waits for its price
	if order.Kind().IsPending() {
		return h.placeOrder(ctx, res, e, order, decision.Rule.Name)
	}

	e.Lock()

	if t, ok := h.QuoteFor(r, req.Symbol); ok {
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
		price = tick.OpenPrice(order.Kind().IsBuy())
	}
	price = NormalisePrice(price, r.Digits)

	fill := h.Execute(e, order, r, price, Now())

	h.bookFill(e, order, fill, r)

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
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	e.Lock()

	o, ok := e.Orders[req.OrderId]
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}
	if !o.State.IsLive() {
		e.Unlock()
		return h.refuse(res, model.RetTradeFrozen, "")
	}
	if o.Kind().IsMarket() {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "a market order cannot be modified")
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.QuoteFor(r, o.Symbol)
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
	if req.TypeTime > 0 || req.ExpiryAt > 0 {
		want.TypeTime, want.TimeExpiration = req.TypeTime, req.ExpiryAt
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
		Kind:      model.RouteFlags_modify,
		Order:     &want,
		Entry:     e,
		Rules:     r,
		Tick:      tick,
		Gapped:    h.Quotes.Gapped(o.Symbol),
		Deviation: h.deviation(&want, tick, r),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, &want, model.OrderState_request_modify)
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

	h.PublishWS(model.SubjectAccountOrders(saved.Login), model.EventOrderCreate, &saved)

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
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	e.Lock()

	o, ok := e.Orders[req.OrderId]
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}
	if !o.State.IsLive() {
		e.Unlock()
		return h.refuse(res, model.RetTradeFrozen, "")
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, _ := h.QuoteFor(r, o.Symbol)

	decision := h.Route(&Request{
		Kind: model.RouteFlags_remove, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(o.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.OrderState_request_cancel)
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
		Kind: model.RouteFlags_activate, Order: o, Entry: e, Rules: r, Tick: t,
		Gapped: h.Quotes.Gapped(o.Symbol),
	})

	if !decision.Executes() {
		// cancel order is the rule that exists precisely for this.
		if decision.Action == model.RouteAction_cancel_order {
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
	if decision.AtMarket() && o.Kind() != model.OrderType_buy_limit && o.Kind() != model.OrderType_sell_limit {
		price = t.OpenPrice(o.Kind().IsBuy())
	}
	price = NormalisePrice(price, r.Digits)

	// the fill is a market order of the same side and size, against the pending order's ticket
	fillOrder := *o
	fillOrder.Type = model.OrderType_buy
	if !o.Kind().IsBuy() {
		fillOrder.Type = model.OrderType_sell
	}
	fillOrder.Reason = model.OrderReason_client

	fill := h.Execute(e, &fillOrder, r, price, Now())

	delete(e.Orders, o.OrderId)
	o.State = model.OrderState_filled
	o.TimeDone = Now()
	o.PriceCurrent = price

	h.bookFill(e, &fillOrder, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save an activation",
			"login", o.Login, "order", o.OrderId, "error", err.Error())
		return err
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "pending order filled",
		"login", o.Login, "order", o.OrderId, "type", o.Type,
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

// placeOrder puts a working order on the book and leaves it there.
func (h *Handler) placeOrder(ctx context.Context, res *model.TradeResult, e *book.Entry,
	o *model.Order, rule string) *model.TradeResult {
	o.State = model.OrderState_placed

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
	h.PublishWS(model.SubjectAccountOrders(o.Login), model.EventOrderCreate, o)

	// a working order can reserve margin of its own, so the account changed even with no deal
	h.CalculateAccountMarginsAndProfits(ctx, e)

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

// bookFill puts the fill onto the account and recomputes the money.
func (h *Handler) bookFill(e *book.Entry, o *model.Order, f *Fill, r *settings.Rules) {
	h.ChargeCommission(e, f, r)

	for _, p := range f.Closed {
		delete(e.Positions, p.PositionId)
		h.SettleSwap(e, p)
	}

	h.creditRealised(e, f.Profit)

	if f.Opened != nil {
		// the id is filled in when the row is written.
		f.Opened.PositionId = h.nextTempId()
		e.Positions[f.Opened.PositionId] = f.Opened
	}

	o.State = model.OrderState_filled
	o.TimeDone = Now()
	o.PriceCurrent = f.Price

	h.CalculateAccountMargins(e).Apply(e.Account)
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
		State:          model.OrderState_started,
		Reason:         req.Reason,
		TimeSetup:      now,
		TimeExpiration: req.ExpiryAt,
		Type:           req.Type,
		TypeFill:       req.TypeFill,
		TypeTime:       req.TypeTime,
		PriceOrder:     req.Price,
		PriceTrigger:   req.PriceTrigger,
		PriceCurrent:   t.OpenPrice(req.Type.IsBuy()),
		PriceSL:        req.PriceSL,
		PriceTP:        req.PriceTP,
		VolumeInitial:  req.Volume,
		VolumeCurrent:  req.Volume,
		ExpertId:       req.ExpertId,
		PositionId:     req.PositionId,
		PositionById:   req.PositionById,
		Comment:        req.Comment,
		RateMargin:     h.RateMargin(r, a, req.Type.IsBuy()),
	}
}

// kindOf is which routing request type this order counts as, which decides what rules see it.
func (h *Handler) kindOf(o *model.Order, r *settings.Rules) model.RouteFlags {
	if o.Kind().IsPending() {
		return model.RouteFlags_pending
	}
	if o.Kind() == model.OrderType_close_by {
		return model.RouteFlags_close_by
	}

	switch r.ExecMode {
	case model.ExecMode_instant:
		return model.RouteFlags_instant
	case model.ExecMode_request:
		return model.RouteFlags_request
	case model.ExecMode_exchange:
		return model.RouteFlags_exchange
	default:
		return model.RouteFlags_market
	}
}

// deviation is how far the requested price is from the market, in points.
func (h *Handler) deviation(o *model.Order, t model.Tick, r *settings.Rules) float64 {
	if o.PriceOrder <= 0 || r.Point <= 0 {
		return 0
	}

	if o.Kind().IsBuy() {
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
	case model.RouteAction_reject:
		return h.refuse(res, model.RetTradeRejected, d.Reason)
	case model.RouteAction_requote:
		return h.refuse(res, model.RetTradeRequote, "")
	case model.RouteAction_dealer, model.RouteAction_dealer_online:
		o.State = state
		return h.SendDealing(req, o, d.Dealers, d.Rule.RoutingId)
	case model.RouteAction_cancel_order:
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

	if amount > 0 && ok && g.MarginFreeProfit == int32(model.FreeMarginProfitMode_day_profit_loss) {
		e.Account.BlockedProfit += amount
		return
	}

	e.Account.Balance += amount
}
