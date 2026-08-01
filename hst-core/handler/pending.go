package handler

import (
	"context"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Pending orders: the ones waiting for a price.
//
// A limit waits for the price to come to it — a buy limit sits below the market and fills when
// the market falls to it. A stop waits for the price to run away — a buy stop sits above and
// fills when the market rises through it. A stop limit does both: the stop level turns it into
// a limit order, which then waits again.
//
// They are checked on every tick for their symbol, in the same pass that prices positions.

// pendingHit is an order the price reached.
type pendingHit struct {
	order   *model.Order
	toLimit bool // a stop limit that has become a limit rather than filling
}

// pendingFor collects the orders this tick triggered. The account's lock must be held.
func (h *Handler) pendingFor(e *book.Entry, t model.Tick) []pendingHit {
	var hits []pendingHit

	for _, o := range e.Orders {
		if o.Symbol != t.Symbol || !model.OrderState(o.State).Live() {
			continue
		}

		kind := o.Kind()
		if !kind.Pending() {
			continue
		}

		// an order fills at the price its own side pays
		price := t.OpenPrice(kind.Buy())

		switch kind {
		case model.OrderBuyLimit:
			if price <= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderSellLimit:
			if price >= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderBuyStop:
			if price >= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderSellStop:
			if price <= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}

		// a stop limit does not fill when its stop is reached; it becomes a limit order at the
		// price the client asked for and waits again
		case model.OrderBuyStopLimit:
			if price >= o.PriceTrigger {
				hits = append(hits, pendingHit{order: o, toLimit: true})
			}
		case model.OrderSellStopLimit:
			if price <= o.PriceTrigger {
				hits = append(hits, pendingHit{order: o, toLimit: true})
			}
		}
	}

	return hits
}

// activate fills a pending order, or turns a stop limit into a limit.
func (h *Handler) activate(ctx context.Context, e *book.Entry, hit pendingHit, t model.Tick) {
	o := hit.order

	e.Lock()

	// another tick may have taken it already
	if _, still := e.Orders[o.OrderId]; !still {
		e.Unlock()
		return
	}

	r, ok := h.Settings.For(e.Account.Group, o.Symbol)
	if !ok {
		e.Unlock()
		return
	}

	if hit.toLimit {
		h.stopLimitToLimit(ctx, e, o, r)
		return
	}

	// the rules see an activation as its own kind of request, so a broker can hold one during a
	// gap instead of filling at whatever the jump produced
	decision := h.Route(&Request{
		Kind: model.RouteActivate, Order: o, Entry: e, Rules: r, Tick: t,
		Gapped: h.Quotes.Gapped(o.Symbol),
	})

	if !decision.Executes() {
		// cancel order is the rule that exists precisely for this: an order that would otherwise
		// keep triggering on every tick and being refused
		if decision.Action == model.ActionCancelOrder {
			h.cancelOrder(ctx, e, o, "deleted [by routing rule]")
			return
		}

		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "an activation was not admitted by any rule",
			"login", o.Login, "order", o.OrderId)

		return
	}

	// the account may have moved since the order was placed, so the margin is checked again at
	// the moment it fires rather than only when it was accepted
	if code := h.checkMoney(e, o, r, t); !code.OK() {
		h.cancelOrder(ctx, e, o, "canceled, not enough money")
		return
	}

	// a limit fills at its own price; anything else fills at the market
	price := o.PriceOrder
	if decision.AtMarket() && o.Kind() != model.OrderBuyLimit && o.Kind() != model.OrderSellLimit {
		price = t.OpenPrice(o.Kind().Buy())
	}
	price = NormalisePrice(price, r.Digits)

	// the fill is a market order of the same side and size
	fillOrder := *o
	fillOrder.Type = int32(model.OrderBuy)
	if !o.Kind().Buy() {
		fillOrder.Type = int32(model.OrderSell)
	}
	fillOrder.Reason = int32(model.ReasonClient)

	fill := h.Execute(e, &fillOrder, r, price, Now())

	// the pending order is done; it becomes history
	delete(e.Orders, o.OrderId)
	o.State = int32(model.StateFilled)
	o.TimeDone = Now()
	o.PriceCurrent = price

	h.settle(e, &fillOrder, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.finishActivation(ctx, e, o, &fillOrder, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save an activation",
			"login", o.Login, "order", o.OrderId, "error", err.Error())
		return
	}

	h.publishTrade(e, o, fill, &account)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "pending order filled",
		"login", o.Login, "order", o.OrderId, "type", model.OrderType(o.Type),
		"price", price)
}

// stopLimitToLimit turns a triggered stop limit into the limit order it was always going to
// become. The account's lock is held on entry and released here.
func (h *Handler) stopLimitToLimit(ctx context.Context, e *book.Entry, o *model.Order,
	r *settings.Rules) {
	if o.Kind() == model.OrderBuyStopLimit {
		o.Type = int32(model.OrderBuyLimit)
	} else {
		o.Type = int32(model.OrderSellLimit)
	}

	o.ActivationMode = model.ActivationStopLimit
	o.ActivationTime = Now()
	o.ActivationPrice = o.PriceTrigger

	login, id, kind := o.Login, o.OrderId, o.Kind()
	e.Unlock()

	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.orders SET type = $1, activation_mode = $2, activation_time = $3,
		        activation_price = $4, date_modified = $3
		  WHERE order_id = $5`,
		int32(kind), int32(model.ActivationStopLimit), o.ActivationTime,
		o.ActivationPrice, id); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not convert a stop limit",
			"login", login, "order", id, "error", err.Error())
		return
	}

	h.publish(subjectOrders(login), "order", o)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "stop limit became a limit",
		"login", login, "order", id, "price", o.PriceOrder)
}

// finishActivation writes the filled pending order and everything it produced.
func (h *Handler) finishActivation(ctx context.Context, e *book.Entry, pending, fillOrder *model.Order,
	f *Fill, a *model.Account) error {
	// the pending order already has a row; mark it done rather than inserting another
	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.orders SET state = $1, time_done = $2, price_current = $3, date_modified = $2
		  WHERE order_id = $4`,
		int32(model.StateFilled), pending.TimeDone, pending.PriceCurrent, pending.OrderId); err != nil {
		return err
	}

	// the deals and positions are written against the pending order's own ticket
	fillOrder.OrderId = pending.OrderId

	return h.saveFill(ctx, e, fillOrder, f, a)
}

// cancelOrder takes a working order off. The account's lock is held on entry.
func (h *Handler) cancelOrder(ctx context.Context, e *book.Entry, o *model.Order, comment string) {
	delete(e.Orders, o.OrderId)

	o.State = int32(model.StateCanceled)
	o.TimeDone = Now()
	if comment != "" {
		o.Comment = comment
	}

	login, id, symbol := o.Login, o.OrderId, o.Symbol
	e.Unlock()

	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.orders SET state = $1, time_done = $2, comment = $3, date_modified = $2
		  WHERE order_id = $4`,
		int32(model.StateCanceled), o.TimeDone, o.Comment, id); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not cancel an order",
			"login", login, "order", id, "error", err.Error())
		return
	}

	h.unwatchIfLast(e, symbol)
	h.publish(subjectOrders(login), "order_canceled", o)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "order canceled",
		"login", login, "order", id, "comment", comment)
}

// expireOrders takes off the orders whose time has run out. Called on the tick path, so an
// order expires the moment its symbol next quotes rather than waiting for a sweep.
func (h *Handler) expireOrders(ctx context.Context, e *book.Entry, symbol string) {
	now := Now()

	e.Lock()

	var expired []*model.Order
	for _, o := range e.Orders {
		if o.Symbol != symbol || !model.OrderState(o.State).Live() {
			continue
		}
		if o.TimeExpiration > 0 && o.TimeExpiration <= now {
			expired = append(expired, o)
		}
	}

	e.Unlock()

	for _, o := range expired {
		e.Lock()
		if _, still := e.Orders[o.OrderId]; !still {
			e.Unlock()
			continue
		}
		h.cancelOrder(ctx, e, o, "expired")
	}
}
