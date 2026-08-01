package handler

import (
	"math"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

// The checks a trade request must pass before the routing rules ever see it.
//
// Ordered cheapest first, and nothing here reads the database: everything it needs is already
// in memory. Each check answers with a return code rather than a bare false, so the client is
// told which rule it broke.
//
// These are the broker's own limits — what the group allows, what the instrument allows, what
// the account can afford. The routing rules that run afterwards are a separate question: what
// the broker wants to do about a request that is otherwise legal.

// Check runs the whole chain and returns the first refusal.
func (h *Handler) Check(e *book.Entry, o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	if code := h.checkAccount(e); !code.OK() {
		return code
	}
	if code := h.checkSymbol(o, r); !code.OK() {
		return code
	}
	if code := h.checkQuote(r, t); !code.OK() {
		return code
	}
	if code := h.checkVolume(o, r); !code.OK() {
		return code
	}
	if code := h.checkLimits(e, o, r); !code.OK() {
		return code
	}
	if code := h.checkFilling(o, r); !code.OK() {
		return code
	}
	if code := h.checkExpiry(o, r); !code.OK() {
		return code
	}
	if code := h.checkStops(o, r, t); !code.OK() {
		return code
	}
	if code := h.checkMoney(e, o, r, t); !code.OK() {
		return code
	}

	return model.RetOK
}

// checkAccount refuses a login that may not trade at all.
func (h *Handler) checkAccount(e *book.Entry) model.RetCode {
	if e == nil || e.Account == nil {
		return model.RetNotFound
	}

	// The trading bits are refusals, not permissions: an ordinary account carries none of them.
	// Reading trade_disabled as "may trade" would refuse everybody and let only the accounts a
	// manager had explicitly stopped through.
	const (
		enabled       = 0x0001
		tradeDisabled = 0x0004
		investor      = 0x0008
		readOnly      = 0x0200
	)

	if e.Account.Rights&enabled == 0 {
		return model.RetAuthDisabled
	}
	if e.Account.Rights&tradeDisabled != 0 {
		return model.RetTradeDisabled
	}
	// an investor password sees everything and trades nothing
	if e.Account.Rights&(investor|readOnly) != 0 {
		return model.RetTradeDisabled
	}

	return model.RetOK
}

// checkSymbol refuses an instrument the group may not trade, or may not trade this way.
func (h *Handler) checkSymbol(o *model.Order, r *settings.Rules) model.RetCode {
	if r == nil {
		return model.RetTradeBadSymbol
	}

	switch {
	case r.TradeMode == model.TradeDisabled:
		return model.RetTradeDisabled
	case r.TradeMode.CloseOnly():
		return model.RetTradeCloseOnly
	case o.Kind().Buy() && !r.TradeMode.AllowsBuy():
		return model.RetTradeDisabled
	case !o.Kind().Buy() && !r.TradeMode.AllowsSell():
		return model.RetTradeDisabled
	}

	return model.RetOK
}

// checkQuote refuses a trade with no usable price behind it.
func (h *Handler) checkQuote(r *settings.Rules, t model.Tick) model.RetCode {
	if !t.Ok() {
		return model.RetTradeNoQuotes
	}

	// a quote older than the symbol allows is not a price any more
	if r.Symbol != nil && r.Symbol.QuotesTime > 0 {
		age := time.Since(time.Unix(0, t.Time))
		if age > time.Duration(r.Symbol.QuotesTime)*time.Second {
			return model.RetTradeNoQuotes
		}
	}

	return model.RetOK
}

// checkVolume applies the instrument's size limits.
func (h *Handler) checkVolume(o *model.Order, r *settings.Rules) model.RetCode {
	v := o.VolumeCurrent

	if v <= 0 {
		return model.RetTradeInvalidVolume
	}
	if r.VolumeMin > 0 && v < r.VolumeMin {
		return model.RetTradeInvalidVolume
	}
	if r.VolumeMax > 0 && v > r.VolumeMax {
		return model.RetTradeInvalidVolume
	}

	// the volume must sit on a step boundary above the minimum
	if r.VolumeStep > 0 && (v-r.VolumeMin)%r.VolumeStep != 0 {
		return model.RetTradeInvalidVolume
	}

	return model.RetOK
}

// checkLimits applies what the group caps: how many orders, how many positions, how much
// volume on one symbol.
func (h *Handler) checkLimits(e *book.Entry, o *model.Order, r *settings.Rules) model.RetCode {
	g := r.Group

	if g.LimitOrders > 0 && int32(len(e.Orders)) >= g.LimitOrders {
		return model.RetTradeTooManyOrder
	}
	if g.LimitPositions > 0 && int32(len(e.Positions)) >= g.LimitPositions {
		return model.RetTradeTooManyOrder
	}

	// the instrument's own cap on how much may be held at once, across everything on it
	if r.VolumeLimit > 0 {
		var held int64
		for _, p := range e.Positions {
			if p.Symbol == o.Symbol {
				held += p.Volume
			}
		}
		if held+o.VolumeCurrent > r.VolumeLimit {
			return model.RetTradeMaxVolume
		}
	}

	return model.RetOK
}

// checkFilling refuses a filling policy the instrument does not offer.
func (h *Handler) checkFilling(o *model.Order, r *settings.Rules) model.RetCode {
	if r.FillFlags == 0 {
		return model.RetOK
	}

	var want int32

	switch model.Filling(o.TypeFill) {
	case model.FillFOK:
		want = model.FillFlagFOK
	case model.FillIOC:
		want = model.FillFlagIOC
	case model.FillReturn:
		want = model.FillFlagReturn
	case model.FillBOC:
		want = model.FillFlagBOC
	}

	if r.FillFlags&want == 0 {
		return model.RetTradeFillPolicy
	}

	return model.RetOK
}

// checkExpiry refuses an expiry type the instrument does not offer, or one already in the past.
func (h *Handler) checkExpiry(o *model.Order, r *settings.Rules) model.RetCode {
	if r.ExpirFlags != 0 {
		var want int32

		switch model.Expiry(o.TypeTime) {
		case model.ExpiryGTC:
			want = model.ExpirFlagGTC
		case model.ExpiryDay:
			want = model.ExpirFlagDay
		case model.ExpirySpecified:
			want = model.ExpirFlagSpecified
		case model.ExpirySpecifiedDay:
			want = model.ExpirFlagSpecifiedDay
		}

		if r.ExpirFlags&want == 0 {
			return model.RetTradeExpiration
		}
	}

	specified := model.Expiry(o.TypeTime) == model.ExpirySpecified ||
		model.Expiry(o.TypeTime) == model.ExpirySpecifiedDay

	if specified {
		if o.TimeExpiration == 0 || o.TimeExpiration <= time.Now().UnixNano() {
			return model.RetTradeExpiration
		}
	}

	return model.RetOK
}

// checkStops keeps stop loss, take profit and a pending order's price far enough from the
// market, using the instrument's stops level.
func (h *Handler) checkStops(o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	if r.StopsLevel <= 0 || r.Point <= 0 {
		return model.RetOK
	}

	minDistance := float64(r.StopsLevel) * r.Point
	buy := o.Kind().Buy()

	// the price the levels are measured against: where this side would close
	market := t.ClosePrice(buy)

	if o.PriceSL > 0 {
		if buy && market-o.PriceSL < minDistance {
			return model.RetTradeInvalidStops
		}
		if !buy && o.PriceSL-market < minDistance {
			return model.RetTradeInvalidStops
		}
	}

	if o.PriceTP > 0 {
		if buy && o.PriceTP-market < minDistance {
			return model.RetTradeInvalidStops
		}
		if !buy && market-o.PriceTP < minDistance {
			return model.RetTradeInvalidStops
		}
	}

	// a pending order must also sit away from the current price, or it would fill at once
	if o.Kind().Pending() && o.PriceOrder > 0 {
		if math.Abs(t.OpenPrice(buy)-o.PriceOrder) < minDistance {
			return model.RetTradeInvalidStops
		}
	}

	return model.RetOK
}

// checkMoney refuses a trade the account cannot cover.
//
// Only opening costs money. A trade that closes or shrinks a position frees margin rather than
// taking it, so it is allowed even when free margin is already negative.
func (h *Handler) checkMoney(e *book.Entry, o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	if h.closesPosition(e, o, r) {
		return model.RetOK
	}

	price := o.PriceOrder
	if price <= 0 {
		price = t.OpenPrice(o.Kind().Buy())
	}

	need := MarginFor(r, o.Lots(), price, e.Account.Leverage, o.RateMargin)

	money := Settle(e.Account, e.Positions, r.Group.MarginFreeProfit != 0)
	if money.FreeMargin < need {
		return model.RetTradeNoMoney
	}

	return model.RetOK
}

// closesPosition reports whether the order would reduce what is already open rather than add
// to it. Under netting an opposite order shrinks the position; under hedging every order opens
// something new.
func (h *Handler) closesPosition(e *book.Entry, o *model.Order, r *settings.Rules) bool {
	if model.MarginMode(r.Group.MarginMode).Hedging() {
		return false
	}

	for _, p := range e.Positions {
		if p.Symbol == o.Symbol && p.Buy() != o.Kind().Buy() {
			return true
		}
	}

	return false
}
