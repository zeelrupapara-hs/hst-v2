package handler

import (
	"hstcore/pkg/logger"
	"math"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

// ValidateOrder runs the whole chain and returns the first refusal.
func (h *Handler) ValidateOrder(e *book.Entry, o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	if code := h.checkAccount(e); !code.OK() {
		return code
	}
	if code := h.checkSymbol(o, r); !code.OK() {
		return code
	}
	if code := h.checkMarket(o, r); !code.OK() {
		return code
	}
	if code := h.checkQuote(r, t); !code.OK() {
		return code
	}
	if code := h.checkOrderFlags(o, r); !code.OK() {
		return code
	}
	if code := h.checkExpert(o, r); !code.OK() {
		return code
	}
	if code := h.checkVolume(o, r); !code.OK() {
		return code
	}
	if code := h.checkHedging(e, o, r); !code.OK() {
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
	if code := h.checkFreeze(o, r, t); !code.OK() {
		return code
	}
	if code := h.checkMoney(e, o, r, t); !code.OK() {
		return code
	}

	return model.RetOK
}

// ValidatePosition checks a change of levels on a position that must still be open.
func (h *Handler) ValidatePosition(e *book.Entry, p *model.Position, req *model.TradeRequest,
	r *settings.Rules, t model.Tick) model.RetCode {
	if code := h.checkAccount(e); !code.OK() {
		return code
	}
	if p == nil || e.Positions[p.PositionId] == nil {
		return model.RetNotFound
	}
	if code := h.checkQuote(r, t); !code.OK() {
		return code
	}
	if r.TradeMode == model.TradeMode_disabled {
		return model.RetTradeDisabled
	}

	// the levels are measured from a position, so the side is the position's own
	level := &model.Order{
		Symbol:     p.Symbol,
		PositionId: p.PositionId,
		Type:       model.OrderType_buy,
		PriceSL:    req.PriceSL,
		PriceTP:    req.PriceTP,
		RateMargin: p.RateMargin,
		ExpertId:   req.ExpertId,
		Reason:     req.Reason,
	}
	if !p.IsBuy() {
		level.Type = model.OrderType_sell
	}

	if code := h.checkOrderFlags(level, r); !code.OK() {
		return code
	}
	if code := h.checkExpert(level, r); !code.OK() {
		return code
	}
	if code := h.checkTrailing(p, level, r); !code.OK() {
		return code
	}
	if code := h.checkStops(level, r, t); !code.OK() {
		return code
	}

	return h.checkFreeze(level, r, t)
}

// checkOrderFlags refuses an order type, or a level, the group does not offer on this instrument.
func (h *Handler) checkOrderFlags(o *model.Order, r *settings.Rules) model.RetCode {
	if r.OrderFlags == 0 {
		return model.RetOK
	}

	if want := model.OrderFlagFor(o.Kind()); want != 0 && r.OrderFlags&want == 0 {
		return model.RetTradeDisabled
	}
	if o.PriceSL > 0 && r.OrderFlags&model.OrderFlagSL == 0 {
		return model.RetTradeInvalidStops
	}
	if o.PriceTP > 0 && r.OrderFlags&model.OrderFlagTP == 0 {
		return model.RetTradeInvalidStops
	}

	return model.RetOK
}

// checkExpert refuses a request an expert advisor sent, when the group does not allow one.
func (h *Handler) checkExpert(o *model.Order, r *settings.Rules) model.RetCode {
	if r.Group.TradeFlags&model.TradeFlagExperts != 0 {
		return model.RetOK
	}
	// routing already reads an expert as a non-zero expert id, so the flag reads it the same way
	if o.ExpertId != 0 || o.Reason == model.OrderReason_expert {
		return model.RetTradeDisabled
	}

	return model.RetOK
}

// checkTrailing refuses an expert moving a stop that is already set, when the group forbids trailing.
func (h *Handler) checkTrailing(p *model.Position, o *model.Order, r *settings.Rules) model.RetCode {
	if r.Group.TradeFlags&model.TradeFlagTrailing != 0 {
		return model.RetOK
	}
	// a trailing stop reaches the engine as an expert dragging an existing stop-loss along
	if p.PriceSL > 0 && o.PriceSL != p.PriceSL &&
		(o.ExpertId != 0 || o.Reason == model.OrderReason_expert) {
		return model.RetTradeInvalidStops
	}

	return model.RetOK
}

// checkHedging refuses a position opposite to one already open, when the group forbids hedging.
func (h *Handler) checkHedging(e *book.Entry, o *model.Order, r *settings.Rules) model.RetCode {
	if r.Group.TradeFlags&model.TradeFlagHedgeProhibit == 0 {
		return model.RetOK
	}
	if !model.MarginMode(r.Group.MarginMode).Hedging() || o.PositionId != 0 {
		return model.RetOK
	}

	for _, p := range e.Positions {
		if p.Symbol == o.Symbol && p.IsBuy() != o.Kind().IsBuy() {
			return model.RetTradeHedgeProhibited
		}
	}

	return model.RetOK
}

// checkFreeze refuses a change too close to the market, where the levels are frozen.
func (h *Handler) checkFreeze(o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	// only an existing order or position freezes; a new one has nothing to protect
	if r.FreezeLevel <= 0 || r.Point <= 0 || (o.OrderId == 0 && o.PositionId == 0) {
		return model.RetOK
	}

	frozen := float64(r.FreezeLevel) * r.Point
	buy := o.Kind().IsBuy()
	market := t.ClosePrice(buy)

	for _, level := range []float64{o.PriceSL, o.PriceTP} {
		if level > 0 && math.Abs(market-level) < frozen {
			return model.RetTradeFrozen
		}
	}

	if o.Kind().IsPending() && o.PriceOrder > 0 &&
		math.Abs(t.OpenPrice(buy)-o.PriceOrder) < frozen {
		return model.RetTradeFrozen
	}

	return model.RetOK
}

// checkExecution applies the instrument's execution mode.
func (h *Handler) checkExecution(o *model.Order, r *settings.Rules,
	t model.Tick) (model.RetCode, model.RouteFlags) {
	kind := h.kindOf(o, r)

	if kind != model.RouteFlags_instant {
		return model.RetOK, kind
	}

	// instant execution above the size the broker fills on the spot is not refused.
	if r.MaxInstantVolume > 0 && o.VolumeCurrent > r.MaxInstantVolume {
		return model.RetOK, model.RouteFlags_request
	}

	// a quote the client answered too late is no longer a price they can be held to
	if r.MaxDeviationTime > 0 &&
		time.Since(time.Unix(0, t.Time)) > time.Duration(r.MaxDeviationTime)*time.Second {
		return model.RetTradeRequote, kind
	}

	if o.PriceOrder > 0 {
		slip := h.deviation(o, t, r)

		if slip > 0 && r.MaxDeviationProfit > 0 && slip > float64(r.MaxDeviationProfit) {
			return model.RetTradeRequote, kind
		}
		if slip < 0 && r.MaxDeviationLoss > 0 && -slip > float64(r.MaxDeviationLoss) {
			return model.RetTradeRequote, kind
		}
	}

	return model.RetOK, kind
}

// checkAccount refuses a login that may not trade at all.
func (h *Handler) checkAccount(e *book.Entry) model.RetCode {
	if e == nil || e.Account == nil {
		return model.RetNotFound
	}

	// The trading bits are refusals, not permissions.
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
	case r.TradeMode == model.TradeMode_disabled:
		return model.RetTradeDisabled
	case r.TradeMode.CloseOnly():
		return model.RetTradeCloseOnly
	case o.Kind().IsBuy() && !r.TradeMode.AllowsBuy():
		return model.RetTradeDisabled
	case !o.Kind().IsBuy() && !r.TradeMode.AllowsSell():
		return model.RetTradeDisabled
	}

	return model.RetOK
}

// checkQuote refuses a trade with no usable price behind it.
func (h *Handler) checkMarket(o *model.Order, r *settings.Rules) model.RetCode {
	dir := model.Direction_in
	if o.PositionId != 0 {
		dir = model.Direction_out
	}

	if !h.IsMarketOpen(r, dir) {
		return model.RetTradeMarketClosed
	}

	return model.RetOK
}

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

// checkLimits applies what the group caps.
func (h *Handler) checkLimits(e *book.Entry, o *model.Order, r *settings.Rules) model.RetCode {
	g := r.Group

	if g.LimitOrders > 0 && len(e.Orders) >= int(g.LimitOrders) {
		return model.RetTradeTooManyOrder
	}
	if g.LimitPositions > 0 && len(e.Positions) >= int(g.LimitPositions) {
		return model.RetTradeTooManyOrder
	}

	var sameSide, total int64
	symbols := make(map[string]bool, len(e.Positions))
	buying := o.Kind().IsBuy()

	for _, p := range e.Positions {
		total += p.Volume
		symbols[p.Symbol] = true
		if p.Symbol == o.Symbol && p.IsBuy() == buying {
			sameSide += p.Volume
		}
	}

	// a working order counts toward the cap too, because it is volume waiting to become a position
	for _, w := range e.Orders {
		if w.Symbol != o.Symbol || w.OrderId == o.OrderId {
			continue
		}
		if !w.Kind().IsPending() || !w.State.IsLive() {
			continue
		}
		if w.Kind().IsBuy() == buying {
			sameSide += w.VolumeCurrent
		}
	}

	// the instrument's cap on how much may be held in one direction, positions and orders together
	if r.VolumeLimit > 0 && sameSide+o.VolumeCurrent > r.VolumeLimit {
		return model.RetTradeMaxVolume
	}

	// limit_positions_volume caps the whole book in lots, not in money, despite the column's name
	if g.LimitPositionsValue > 0 &&
		model.Lots(total+o.VolumeCurrent) > g.LimitPositionsValue {
		return model.RetTradeMaxVolume
	}

	// the group's cap on how many different instruments may be held at once
	if g.LimitSymbols > 0 && !symbols[o.Symbol] && len(symbols) >= int(g.LimitSymbols) {
		return model.RetTradeMaxVolume
	}

	return model.RetOK
}

// checkFilling refuses a filling policy the instrument does not offer.
func (h *Handler) checkFilling(o *model.Order, r *settings.Rules) model.RetCode {
	// book or cancel only ever sits in the book, so an order that would fill on entry is refused
	if o.TypeFill == model.OrderFilling_boc && o.Kind().IsMarket() {
		return model.RetTradeFillPolicy
	}

	if r.FillFlags == 0 {
		return model.RetOK
	}

	var want int32

	switch o.TypeFill {
	case model.OrderFilling_fok:
		want = model.FillFlagFOK
	case model.OrderFilling_ioc:
		want = model.FillFlagIOC
	case model.OrderFilling_return:
		want = model.FillFlagReturn
	case model.OrderFilling_boc:
		want = model.FillFlagBOC
	}

	if r.FillFlags&want == 0 {
		return model.RetTradeFillPolicy
	}

	return model.RetOK
}

// checkExpiry refuses an expiry type the instrument does not offer, or one already in the past.
func (h *Handler) checkExpiry(o *model.Order, r *settings.Rules) model.RetCode {
	// a group that does not allow expiry at all leaves good-till-cancelled as the only choice
	if r.Group.TradeFlags&model.TradeFlagExpiration == 0 &&
		o.TypeTime != model.OrderTime_gtc {
		return model.RetTradeExpiration
	}

	if r.ExpirFlags != 0 {
		var want int32

		switch o.TypeTime {
		case model.OrderTime_gtc:
			want = model.ExpirFlagGTC
		case model.OrderTime_day:
			want = model.ExpirFlagDay
		case model.OrderTime_specified:
			want = model.ExpirFlagSpecified
		case model.OrderTime_specified_day:
			want = model.ExpirFlagSpecifiedDay
		}

		if r.ExpirFlags&want == 0 {
			return model.RetTradeExpiration
		}
	}

	specified := o.TypeTime == model.OrderTime_specified ||
		o.TypeTime == model.OrderTime_specified_day

	if specified {
		if o.TimeExpiration == 0 || o.TimeExpiration <= time.Now().UnixNano() {
			return model.RetTradeExpiration
		}
	}

	return model.RetOK
}

// checkStops keeps stop loss, take profit and a pending order's price far enough from the market, using.
func (h *Handler) checkStops(o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	if r.StopsLevel <= 0 || r.Point <= 0 {
		return model.RetOK
	}

	minDistance := float64(r.StopsLevel) * r.Point
	buy := o.Kind().IsBuy()

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
	if o.Kind().IsPending() && o.PriceOrder > 0 {
		if math.Abs(t.OpenPrice(buy)-o.PriceOrder) < minDistance {
			return model.RetTradeInvalidStops
		}
	}

	return model.RetOK
}

// checkMoney refuses a trade the account cannot cover.
func (h *Handler) checkMoney(e *book.Entry, o *model.Order, r *settings.Rules, t model.Tick) model.RetCode {
	opening := h.openingVolume(e, o, r)
	if opening <= 0 {
		return model.RetOK
	}

	price := o.PriceOrder
	if price <= 0 {
		price = t.OpenPrice(o.Kind().IsBuy())
	}

	// only the part that opens exposure is charged, and the order's own type decides the
	// multiplier, so a pending type rated zero costs nothing
	need := MarginForType(r, model.Lots(opening), price, e.Account.Leverage, o.RateMargin, o.Kind(), false)

	money := h.CalculateAccountMargins(e)
	if money.FreeMargin < need {
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "money check refused",
			"login", e.Account.Login, "need", need, "free", money.FreeMargin,
			"balance", money.Balance, "margin", money.Margin, "leverage", e.Account.Leverage,
			"rate", o.RateMargin)
		return model.RetTradeNoMoney
	}

	return model.RetOK
}

// openingVolume is how much of an order opens exposure rather than closing what is already there.
//
// Only the closing part is free of margin. An order larger than the position it offsets still
// opens the difference, and that difference has to be paid for.
func (h *Handler) openingVolume(e *book.Entry, o *model.Order, r *settings.Rules) int64 {
	if model.MarginMode(r.Group.MarginMode).Hedging() {
		return o.VolumeCurrent
	}

	var opposite int64
	for _, p := range e.Positions {
		if p.Symbol == o.Symbol && p.IsBuy() != o.Kind().IsBuy() {
			opposite += p.Volume
		}
	}

	if opposite >= o.VolumeCurrent {
		return 0
	}

	return o.VolumeCurrent - opposite
}
