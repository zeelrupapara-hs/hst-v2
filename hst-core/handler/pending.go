package handler

import (
	wire "hstmodel"

	"context"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// pendingHit is an order the price reached.
type pendingHit struct {
	order   *model.Order
	toLimit bool // a stop limit that has become a limit rather than filling
}

// pendingHits collects the orders this tick triggered. The account's lock must be held.
func (h *Handler) pendingHits(e *book.Entry, t model.Tick) []pendingHit {
	var hits []pendingHit

	for _, o := range e.Orders {
		if o.Symbol != t.Symbol || !o.State.IsLive() {
			continue
		}

		kind := o.Kind()
		if !kind.IsPending() {
			continue
		}

		// an order fills at the price its own side pays
		price := t.OpenPrice(kind.IsBuy())

		switch kind {
		case model.OrderType_buy_limit:
			if price <= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderType_sell_limit:
			if price >= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderType_buy_stop:
			if price >= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}
		case model.OrderType_sell_stop:
			if price <= o.PriceOrder {
				hits = append(hits, pendingHit{order: o})
			}

		// a stop limit does not fill when its stop is reached.
		case model.OrderType_buy_stop_limit:
			if price >= o.PriceTrigger {
				hits = append(hits, pendingHit{order: o, toLimit: true})
			}
		case model.OrderType_sell_stop_limit:
			if price <= o.PriceTrigger {
				hits = append(hits, pendingHit{order: o, toLimit: true})
			}
		}
	}

	return hits
}

// cookStopLimit turns a triggered stop limit into the limit order it was always going to become.
func (h *Handler) cookStopLimit(ctx context.Context, e *book.Entry, o *model.Order,
	r *settings.Rules) {
	if o.Kind() == model.OrderType_buy_stop_limit {
		o.Type = model.OrderType_buy_limit
	} else {
		o.Type = model.OrderType_sell_limit
	}

	o.ActivationMode = model.ActivationStopLimit
	o.ActivationTime = Now()
	o.ActivationPrice = o.PriceTrigger

	saved := *o
	e.Unlock()

	if err := h.writeOrder(ctx, &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not convert a stop limit",
			"login", saved.Login, "order", saved.OrderId, "error", err.Error())
		return
	}

	h.PublishWS(model.SubjectAccountOrders(saved.Login), wire.EventOrderCreate, &saved)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "stop limit became a limit",
		"login", saved.Login, "order", saved.OrderId, "price", saved.PriceOrder)
}

// removeOrder takes a working order off. The account's lock is held on entry.
func (h *Handler) removeOrder(ctx context.Context, e *book.Entry, o *model.Order, comment string) {
	delete(e.Orders, o.OrderId)

	o.State = model.OrderState_canceled
	o.TimeDone = Now()
	if comment != "" {
		o.Comment = comment
	}

	saved := *o
	e.Unlock()

	if err := h.writeOrder(ctx, &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not cancel an order",
			"login", saved.Login, "order", saved.OrderId, "error", err.Error())
		return
	}

	h.unwatchIfLast(e, saved.Symbol)
	h.PublishWS(model.SubjectAccountOrders(saved.Login), wire.EventOrderCancel, &saved)

	// whatever the order was reserving goes back
	h.CalculateAccountMarginsAndProfits(ctx, e)

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "order canceled",
		"login", saved.Login, "order", saved.OrderId, "comment", comment)
}

// ExpireOrders takes off the orders whose time has run out.
func (h *Handler) ExpireOrders(ctx context.Context, e *book.Entry, symbol string) {
	now := Now()

	e.Lock()

	var expired []*model.Order
	for _, o := range e.Orders {
		if o.Symbol != symbol || !o.State.IsLive() {
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
		h.removeOrder(ctx, e, o, "expired")
	}
}
