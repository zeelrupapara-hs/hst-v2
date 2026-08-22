package handler

import (
	"context"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Fill is what a completed order did.
type Fill struct {
	Deals   []*model.Deal
	Opened  *model.Position
	Closed  []*model.Position
	Changed []*model.Position
	Profit  float64
	Price   float64
	RetCode model.RetCode
}

// trigger is a level the price crossed.
type trigger struct {
	position *model.Position
	reason   model.OrderReason
}

// Execute turns an order into deals and positions on the account.
func (h *Handler) Execute(e *book.Entry, o *model.Order, r *settings.Rules, price float64,
	now int64) *Fill {
	f := &Fill{Price: price, RetCode: model.RetOK}

	// an order naming a ticket takes that ticket off, whatever the margin mode.
	if p := h.positionById(e, o.PositionId); p != nil {
		h.refreshRateProfit(e, r, p)
		if o.VolumeCurrent < p.Volume {
			h.reducePosition(f, e, p, o, r, price, now)
		} else {
			h.closeInto(f, e, p, o, r, price, now)
		}
		return f
	}

	if model.MarginMode(r.Group.MarginMode).Hedging() {
		h.NewPosition(f, e, o, r, price, now)
		return f
	}

	h.netInto(f, e, o, r, price, now)

	return f
}

// NewPosition starts a position.
func (h *Handler) NewPosition(f *Fill, e *book.Entry, o *model.Order, r *settings.Rules,
	price float64, now int64) {
	side := model.PositionAction_buy
	if !o.Kind().IsBuy() {
		side = model.PositionAction_sell
	}

	p := &model.Position{
		Login:           o.Login,
		Dealer:          o.Dealer,
		Symbol:          o.Symbol,
		Action:          side,
		Digits:          r.Digits,
		DigitsCurrency:  e.Account.CurrencyDigits,
		Reason:          positionReason(o.Reason),
		ContractSize:    r.ContractSize,
		TimeCreate:      now,
		TimeUpdate:      now,
		PriceOpen:       price,
		PriceCurrent:    price,
		PriceSL:         o.PriceSL,
		PriceTP:         o.PriceTP,
		Volume:          o.VolumeCurrent,
		RateMargin:      o.RateMargin,
		RateProfit:      h.RateProfit(r, e.Account, o.Kind().IsBuy()),
		ExpertId:        o.ExpertId,
		Comment:         o.Comment,
		ActivationFlags: o.ActivationFlags,
	}

	p.Margin = MarginForPosition(r, p, price, e.Account.Leverage)

	f.Opened = p
	f.Deals = append(f.Deals, h.MakeDealIn(o, r, e, price, o.VolumeCurrent, 0, now))
}

// UpdatePosition moves a position's stop loss and take profit, and nothing else.
func (h *Handler) UpdatePosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	p := h.positionById(e, req.PositionId)
	if p == nil {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	r, ok := h.Settings.For(e.Account.Group, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.QuoteFor(r, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	if code := h.ValidatePosition(e, p, req, r, tick); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	// the rules see a level change as its own kind of request, and can strip a level off it
	o := h.closingOrder(p, r, req.Reason, p.Volume, "")
	o.PriceSL, o.PriceTP = req.PriceSL, req.PriceTP

	decision := h.Route(&Request{
		Kind: model.RouteFlags_sltp, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.OrderState_request_add)
	}

	before := *p

	p.PriceSL, p.PriceTP = o.PriceSL, o.PriceTP
	p.TimeUpdate = Now()

	saved := *p
	e.Unlock()

	if err := h.SavePositionAndPublish(ctx, entryGroup(e), &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not modify a position",
			"login", saved.Login, "position", saved.PositionId, "error", err.Error())
		e.Lock()
		*p = before
		e.Unlock()
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.PositionId = saved.PositionId
	res.Volume = saved.Volume
	res.Rule = decision.Rule.Name

	return res
}

// ClosePosition is the client's own close, whole or by volume.
func (h *Handler) ClosePosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	p := h.positionById(e, req.PositionId)
	if p == nil {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	r, ok := h.Settings.For(e.Account.Group, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	tick, ok := h.QuoteFor(r, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	if code := h.checkAccount(e); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	if !h.firstInLine(e, p, r) {
		e.Unlock()
		return h.refuse(res, model.RetTradeCloseOrderExist, "")
	}

	o := h.closingOrder(p, r, req.Reason, req.Volume, req.Comment)
	o.Dealer = req.Dealer
	o.PriceOrder = req.Price

	// a partial close has to leave a volume the instrument can still hold
	if req.Volume > 0 && req.Volume < p.Volume {
		if code := h.checkVolume(o, r); !code.OK() {
			e.Unlock()
			return h.refuse(res, code, "")
		}
	}

	code, kind := h.checkExecution(o, r, tick)
	if !code.OK() {
		e.Unlock()
		res.Bid, res.Ask = tick.Bid, tick.Ask
		return h.refuse(res, code, "")
	}

	decision := h.Route(&Request{
		Kind: kind, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(p.Symbol), Deviation: h.deviation(o, tick, r),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.OrderState_request_add)
	}

	// only instant execution closes where the client asked; every other mode closes at the market
	price := o.PriceOrder
	if kind != model.RouteFlags_instant || decision.AtMarket() || price <= 0 {
		price = tick.ClosePrice(p.IsBuy())
	}
	price = NormalisePrice(price, r.Digits)

	snapshot := e.Snapshot()

	volume := o.VolumeCurrent
	fill := h.Execute(e, o, r, price, Now())

	h.bookFill(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close",
			"login", req.Login, "position", req.PositionId, "error", err.Error())
		h.restore(e, snapshot)
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = o.OrderId
	res.PositionId = req.PositionId
	res.Price = price
	res.Volume = volume
	res.Profit = fill.Profit
	res.Rule = decision.Rule.Name

	if len(fill.Deals) > 0 {
		res.DealId = fill.Deals[0].DealId
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "position closed by the client",
		"login", req.Login, "position", req.PositionId, "volume", model.Lots(volume),
		"price", price, "profit", fill.Profit)

	return res
}

// CloseByPosition takes two opposite positions on one symbol off against each other.
func (h *Handler) CloseByPosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	p, by := h.positionById(e, req.PositionId), h.positionById(e, req.PositionById)
	if p == nil || by == nil || p.PositionId == by.PositionId {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	r, ok := h.Settings.For(e.Account.Group, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeBadSymbol, "")
	}

	if code := h.checkAccount(e); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}
	if !model.MarginMode(r.Group.MarginMode).Hedging() {
		e.Unlock()
		return h.refuse(res, model.RetTradeHedgeProhibited, "")
	}
	if p.Symbol != by.Symbol {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "the two positions are on different symbols")
	}
	if p.IsBuy() == by.IsBuy() {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "the two positions are on the same side")
	}

	tick, ok := h.QuoteFor(r, p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	volume := p.Volume
	if by.Volume < volume {
		volume = by.Volume
	}

	now := Now()

	o := &model.Order{
		Login:          p.Login,
		Dealer:         req.Dealer,
		Symbol:         p.Symbol,
		Digits:         r.Digits,
		DigitsCurrency: e.Account.CurrencyDigits,
		ContractSize:   r.ContractSize,
		State:          model.OrderState_started,
		Reason:         req.Reason,
		TimeSetup:      now,
		Type:           model.OrderType_close_by,
		VolumeInitial:  volume,
		VolumeCurrent:  volume,
		PositionById:   by.PositionId,
		RateMargin:     p.RateMargin,
		Comment:        req.Comment,
	}

	decision := h.Route(&Request{
		Kind: model.RouteFlags_close_by, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.OrderState_request_add)
	}

	h.refreshRateProfit(e, r, p)
	h.refreshRateProfit(e, r, by)

	snapshot := e.Snapshot()

	f := &Fill{Price: by.PriceOpen, RetCode: model.RetOK}
	h.closeAgainst(f, e, p, by, o, r, volume, now)

	// the ticket goes on after the fill, so Execute is not asked to close it a second time
	o.PositionId = p.PositionId

	h.bookFill(e, o, f, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, f, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close by",
			"login", req.Login, "position", req.PositionId, "error", err.Error())
		h.restore(e, snapshot)
		return h.refuse(res, model.RetError, "")
	}

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.OrderId = o.OrderId
	res.PositionId = p.PositionId
	res.Price = by.PriceOpen
	res.Volume = volume
	res.Profit = f.Profit
	res.Rule = decision.Rule.Name

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "position closed by an opposite one",
		"login", req.Login, "position", p.PositionId, "against", by.PositionId,
		"volume", model.Lots(volume), "profit", f.Profit)

	return res
}

// CloseAtMarket is the engine's own close: a stop loss, a take profit or a stop out.
func (h *Handler) CloseAtMarket(ctx context.Context, e *book.Entry, p *model.Position,
	t model.Tick, reason model.OrderReason, kind model.RouteFlags) {
	h.closeAtMarket(ctx, e, p, t, reason, kind, closeComment(reason))
}

// closeAtMarket is CloseAtMarket with the history comment chosen by the caller, which the stop
// out uses to record the level it fired at.
func (h *Handler) closeAtMarket(ctx context.Context, e *book.Entry, p *model.Position,
	t model.Tick, reason model.OrderReason, kind model.RouteFlags, comment string) {
	if !h.lockHeld(e) {
		return
	}

	// it may already have gone: two ticks can pick up the same level
	if _, still := e.Positions[p.PositionId]; !still {
		e.Unlock()
		return
	}

	r, ok := h.Settings.For(e.Account.Group, p.Symbol)
	if !ok {
		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "cannot close: the group has no settings for the symbol",
			"login", p.Login, "group", e.Account.Group, "symbol", p.Symbol)
		return
	}

	o := h.closingOrder(p, r, reason, p.Volume, comment)

	decision := h.Route(&Request{
		Kind: kind, Order: o, Entry: e, Rules: r, Tick: t,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	// the rules can refuse to close, which is a real thing a broker may want during a gap
	if !decision.Executes() {
		e.Unlock()
		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "a close was not admitted by any rule",
			"login", p.Login, "position", p.PositionId, "reason", reason)
		return
	}

	snapshot := e.Snapshot()

	price := NormalisePrice(t.ClosePrice(p.IsBuy()), r.Digits)
	fill := h.Execute(e, o, r, price, Now())

	h.bookFill(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close",
			"login", p.Login, "position", p.PositionId, "error", err.Error())
		h.restore(e, snapshot)
		return
	}

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "position closed",
		"login", p.Login, "position", p.PositionId, "reason", reason,
		"price", price, "profit", fill.Profit)
}

// CookPosition collects the stop losses and take profits this tick crossed.
func (h *Handler) CookPosition(e *book.Entry, t model.Tick) []trigger {
	var hits []trigger

	for _, p := range e.Positions {
		if p.Symbol != t.Symbol {
			continue
		}

		// a position closes at the opposite side to the one it opened on
		price := t.ClosePrice(p.IsBuy())

		sl := p.PriceSL > 0 && p.ActivationFlags&int32(model.ActivationFlags_no_sl) == 0
		tp := p.PriceTP > 0 && p.ActivationFlags&int32(model.ActivationFlags_no_tp) == 0

		var hit trigger

		switch {
		case sl && p.IsBuy() && price <= p.PriceSL:
			hit = trigger{p, model.OrderReason_sl}
		case sl && !p.IsBuy() && price >= p.PriceSL:
			hit = trigger{p, model.OrderReason_sl}
		case tp && p.IsBuy() && price >= p.PriceTP:
			hit = trigger{p, model.OrderReason_tp}
		case tp && !p.IsBuy() && price <= p.PriceTP:
			hit = trigger{p, model.OrderReason_tp}
		default:
			continue
		}

		r, ok := h.Settings.For(e.Account.Group, p.Symbol)
		if !ok {
			continue
		}

		// the first in first out rule binds the server's own closes too: an older position on
		// the instrument means this activation is skipped rather than refused
		if !h.firstInLine(e, p, r) {
			h.Log.Log(logger.TypeTrade, logger.CodeWarn,
				"close prohibited by the first in first out rule, activation skipped",
				"login", p.Login, "position", p.PositionId, "symbol", p.Symbol)
			continue
		}

		// the group can ask for one more margin check before a level closes a position, in case
		// the close would leave what stays open uncovered
		if r.MarginFlags&model.MarginFlagCheckSLTP != 0 && !h.coversAfterClose(e, p) {
			h.Log.Log(logger.TypeTrade, logger.CodeWarn,
				"level activation skipped, the close would leave too little margin",
				"login", p.Login, "position", p.PositionId)
			continue
		}

		hits = append(hits, hit)
	}

	return hits
}

// netInto applies a fill to an account that keeps one position per symbol.
func (h *Handler) netInto(f *Fill, e *book.Entry, o *model.Order, r *settings.Rules,
	price float64, now int64) {
	existing := h.positionFor(e, o.Symbol)

	// nothing open, or the fill is the same way as what is open: open or grow
	if existing == nil {
		h.NewPosition(f, e, o, r, price, now)
		return
	}

	h.refreshRateProfit(e, r, existing)

	if existing.IsBuy() == o.Kind().IsBuy() {
		h.growPosition(f, e, existing, o, r, price, now)
		return
	}

	switch {
	case o.VolumeCurrent < existing.Volume:
		h.reducePosition(f, e, existing, o, r, price, now)
	case o.VolumeCurrent == existing.Volume:
		h.closeInto(f, e, existing, o, r, price, now)
	default:
		h.reversePosition(f, e, existing, o, r, price, now)
	}
}

// growPosition adds to a position and blends the open price.
func (h *Handler) growPosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	total := p.Volume + o.VolumeCurrent

	p.PriceOpen = (p.PriceOpen*float64(p.Volume) + price*float64(o.VolumeCurrent)) / float64(total)
	p.PriceOpen = NormalisePrice(p.PriceOpen, r.Digits)
	p.Volume = total
	p.TimeUpdate = now
	p.Margin = MarginForPosition(r, p, p.PriceOpen, e.Account.Leverage)

	f.Changed = append(f.Changed, p)
	f.Deals = append(f.Deals, h.MakeDealIn(o, r, e, price, o.VolumeCurrent, p.PositionId, now))
}

// reducePosition closes part of a position and books the profit on the part that went.
func (h *Handler) reducePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	closed := o.VolumeCurrent
	profit := ProfitFor(r, p.IsBuy(), model.Lots(closed), p.PriceOpen, price, p.RateProfit)

	// the deal is shaped before the position shrinks, so it carries the swap share of what went
	d := h.MakeDealOut(o, r, e, p, price, closed, now)
	d.Profit = profit

	p.Volume -= closed
	p.Storage -= d.Storage
	p.TimeUpdate = now
	p.Margin = MarginForPosition(r, p, p.PriceOpen, e.Account.Leverage)

	f.Profit += profit
	f.Changed = append(f.Changed, p)
	f.Deals = append(f.Deals, d)
}

// closeInto takes the whole position off.
func (h *Handler) closeInto(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.IsBuy(), p.Lots(), p.PriceOpen, price, p.RateProfit)

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.MakeDealOut(o, r, e, p, price, p.Volume, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)
}

// reversePosition closes what was open and opens the remainder the other way.
func (h *Handler) reversePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.IsBuy(), p.Lots(), p.PriceOpen, price, p.RateProfit)
	remainder := o.VolumeCurrent - p.Volume

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.MakeDealOut(o, r, e, p, price, p.Volume, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)

	// the leftover volume opens a fresh position the other way round
	side := model.PositionAction_buy
	if !o.Kind().IsBuy() {
		side = model.PositionAction_sell
	}

	np := &model.Position{
		Login:           o.Login,
		Dealer:          o.Dealer,
		Symbol:          o.Symbol,
		Action:          side,
		Digits:          r.Digits,
		DigitsCurrency:  e.Account.CurrencyDigits,
		Reason:          positionReason(o.Reason),
		ContractSize:    r.ContractSize,
		TimeCreate:      now,
		TimeUpdate:      now,
		PriceOpen:       price,
		PriceCurrent:    price,
		PriceSL:         o.PriceSL,
		PriceTP:         o.PriceTP,
		Volume:          remainder,
		RateMargin:      o.RateMargin,
		RateProfit:      h.RateProfit(r, e.Account, o.Kind().IsBuy()),
		ExpertId:        o.ExpertId,
		Comment:         o.Comment,
		ActivationFlags: o.ActivationFlags,
	}
	np.Margin = MarginForPosition(r, np, price, e.Account.Leverage)

	f.Opened = np
}

// closeAgainst offsets two opposite positions.
func (h *Handler) closeAgainst(f *Fill, e *book.Entry, p, by *model.Position, o *model.Order,
	r *settings.Rules, volume, now int64) {
	profit := ProfitFor(r, p.IsBuy(), model.Lots(volume), p.PriceOpen, by.PriceOpen, p.RateProfit)
	f.Profit += profit

	d := h.MakeDealOutBy(o, r, e, p, by, by.PriceOpen, volume, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)
	f.Deals = append(f.Deals, h.MakeDealOutBy(o, r, e, by, p, by.PriceOpen, volume, now))

	h.takeOff(f, e, p, r, volume, now)
	h.takeOff(f, e, by, r, volume, now)
}

// takeOff removes volume from a position, closing it when nothing is left.
func (h *Handler) takeOff(f *Fill, e *book.Entry, p *model.Position, r *settings.Rules,
	volume, now int64) {
	if volume >= p.Volume {
		f.Closed = append(f.Closed, p)
		return
	}

	p.Storage -= storageShare(p, volume)
	p.Volume -= volume
	p.TimeUpdate = now
	p.Margin = MarginForPosition(r, p, p.PriceOpen, e.Account.Leverage)

	f.Changed = append(f.Changed, p)
}

// SavePositionAndPublish writes a level change and announces it.
func (h *Handler) SavePositionAndPublish(ctx context.Context, group string, p *model.Position) error {
	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.positions
		    SET price_sl = $1, price_tp = $2, activation_flags = $3, time_update = $4,
		        date_modified = $4, storage = $6
		  WHERE position_id = $5`,
		p.PriceSL, p.PriceTP, p.ActivationFlags, p.TimeUpdate, p.PositionId, p.Storage); err != nil {
		return err
	}

	h.PublishPosition(group, model.EventPositionUpdate, p)

	return nil
}

// positionFor finds the account's position on a symbol. Under netting there is at most one.
func (h *Handler) positionFor(e *book.Entry, symbol string) *model.Position {
	for _, p := range e.Positions {
		if p.Symbol == symbol {
			return p
		}
	}
	return nil
}

// positionById finds one position by its ticket, for the verbs that name one.
// firstInLine reports whether a position may be closed, when the group closes first in first out.
func (h *Handler) firstInLine(e *book.Entry, p *model.Position, r *settings.Rules) bool {
	if r.Group.TradeFlags&model.TradeFlagFIFOClose == 0 {
		return true
	}

	for _, other := range e.Positions {
		if other.Symbol == p.Symbol && other.IsBuy() == p.IsBuy() &&
			other.TimeCreate < p.TimeCreate {
			return false
		}
	}

	return true
}

func (h *Handler) positionById(e *book.Entry, id int64) *model.Position {
	if id <= 0 {
		return nil
	}
	return e.Positions[id]
}

// closingOrder is the order that takes volume off a position.
func (h *Handler) closingOrder(p *model.Position, r *settings.Rules, reason model.OrderReason,
	volume int64, comment string) *model.Order {
	side := model.OrderType_sell
	if !p.IsBuy() {
		side = model.OrderType_buy
	}

	if volume <= 0 || volume > p.Volume {
		volume = p.Volume
	}

	return &model.Order{
		Login:         p.Login,
		Symbol:        p.Symbol,
		Digits:        r.Digits,
		ContractSize:  r.ContractSize,
		State:         model.OrderState_started,
		Reason:        reason,
		TimeSetup:     Now(),
		Type:          side,
		VolumeInitial: volume,
		VolumeCurrent: volume,
		PositionId:    p.PositionId,
		RateMargin:    p.RateMargin,
		Comment:       comment,
	}
}

// positionReason keeps the order-only reasons off a position.
func positionReason(reason model.OrderReason) model.OrderReason {
	switch model.OrderReason(reason) {
	case model.OrderReason_sl, model.OrderReason_tp, model.OrderReason_so:
		return model.OrderReason_client
	}
	return reason
}

// closeComment is what shows against the deal in the client's history.
func closeComment(reason model.OrderReason) string {
	switch reason {
	case model.OrderReason_sl:
		return "[sl]"
	case model.OrderReason_tp:
		return "[tp]"
	case model.OrderReason_so:
		return "[so]"
	}
	return ""
}

// CalcPosition prices every position on a symbol against a new quote.
func (h *Handler) CalcPosition(e *book.Entry, symbol string, t model.Tick) {
	for _, p := range e.Positions {
		if p.Symbol != symbol {
			continue
		}

		r, ok := h.Settings.For(e.Account.Group, symbol)
		if !ok {
			continue
		}

		h.refreshRateProfit(e, r, p)
		p.PriceCurrent = t.ClosePrice(p.IsBuy())
		p.Profit = ProfitFor(r, p.IsBuy(), p.Lots(), p.PriceOpen, p.PriceCurrent, p.RateProfit)
		p.Margin = MarginForPosition(r, p, p.PriceOpen, e.Account.Leverage)
	}
}

// refreshRateProfit values the position at today's conversion rate, keeping the last one known when the cross is missing.
func (h *Handler) refreshRateProfit(e *book.Entry, r *settings.Rules, p *model.Position) {
	if rate, ok := h.crossRate(e.Account.Group, r.CurrencyProfit, e.Account.Currency, p.IsBuy()); ok {
		p.RateProfit = rate
	}
}

// Now is the engine's clock, in epoch nanoseconds.
func Now() int64 { return time.Now().UnixNano() }

// coversAfterClose reports whether the account still covers what stays open once this position
// is gone. Closing one leg of a hedge can raise the margin on the rest.
func (h *Handler) coversAfterClose(e *book.Entry, p *model.Position) bool {
	kept := make(map[int64]*model.Position, len(e.Positions))
	for id, other := range e.Positions {
		if id != p.PositionId {
			kept[id] = other
		}
	}

	if len(kept) == 0 {
		return true
	}

	after := *e.Account
	after.Balance += p.Profit + p.Storage

	free := model.FreeMarginMode_use_pl
	if g, ok := h.Settings.Group(e.Account.Group); ok {
		free = model.FreeMarginMode(g.MarginFreeMode)
	}

	reserved, _ := h.pendingMargin(e)

	return Settle(&after, kept, free, 0, reserved).FreeMargin >= 0
}

// FixPosition writes the deals-derived volume and open price back into the book — the manager's
// correction after the deal history was edited. No order and no deal result; zero volume deletes.
func (h *Handler) FixPosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	if req.Volume <= 0 {
		return h.DeletePosition(ctx, req)
	}

	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	p := h.positionById(e, req.PositionId)
	if p == nil {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	before := *p

	p.Volume = req.Volume
	if req.Price > 0 {
		p.PriceOpen = req.Price
	}
	p.TimeUpdate = Now()

	if r, ok := h.Settings.For(e.Account.Group, p.Symbol); ok {
		if tick, ok := h.QuoteFor(r, p.Symbol); ok {
			price := tick.Bid
			if p.IsBuy() {
				price = tick.Ask
			}
			p.Margin = MarginForPosition(r, p, price, e.Account.Leverage)
		}
	}
	h.CalculateAccountMargins(e).Apply(e.Account)

	saved := *p
	account := *e.Account
	e.Unlock()

	if err := h.savePositionFix(ctx, &saved, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a position fix",
			"login", saved.Login, "position", saved.PositionId, "error", err.Error())
		e.Lock()
		*p = before
		h.CalculateAccountMargins(e).Apply(e.Account)
		e.Unlock()
		return h.refuse(res, model.RetError, "")
	}

	h.PublishPosition(account.Group, model.EventPositionUpdate, &saved)
	h.PublishAccount(&account, nil)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.PositionId = saved.PositionId
	res.Volume = saved.Volume

	h.Log.Log(logger.TypeTrade, logger.CodeAtt, "position fix",
		"login", saved.Login, "position", saved.PositionId,
		"volume", saved.Volume, "price", saved.PriceOpen, "dealer", req.Dealer)

	return res
}

// DeletePosition removes a position without generating an order or a deal — the broker-level
// correction MT5 reserves for a position whose deal history says it should not exist.
func (h *Handler) DeletePosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	if !h.lockHeld(e) {
		return h.refuse(res, model.RetTradeAccountNotFound, "")
	}

	p := h.positionById(e, req.PositionId)
	if p == nil {
		e.Unlock()
		return h.refuse(res, model.RetNotFound, "")
	}

	snapshot := e.Snapshot()
	saved := *p
	delete(e.Positions, p.PositionId)
	h.CalculateAccountMargins(e).Apply(e.Account)
	account := *e.Account
	e.Unlock()

	if err := h.deletePositionRow(ctx, &saved, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not delete a position",
			"login", saved.Login, "position", saved.PositionId, "error", err.Error())
		h.restore(e, snapshot)
		return h.refuse(res, model.RetError, "")
	}

	h.unwatchIfLast(e, saved.Symbol)

	h.PublishPosition(account.Group, model.EventPositionClose, &saved)
	h.PublishAccount(&account, nil)

	res.RetCode = int32(model.RetOK)
	res.Message = model.RetOK.String()
	res.PositionId = saved.PositionId

	h.Log.Log(logger.TypeTrade, logger.CodeAtt, "position deleted",
		"login", saved.Login, "position", saved.PositionId, "dealer", req.Dealer)

	return res
}

// savePositionFix persists a corrected position and the account it re-margins.
func (h *Handler) savePositionFix(ctx context.Context, p *model.Position, a *model.Account) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`UPDATE hst.positions
		    SET volume = $1, volume_ext = $2, price_open = $3, time_update = $4, date_modified = $4
		  WHERE position_id = $5`,
		model.Legacy(p.Volume), p.Volume, p.PriceOpen, p.TimeUpdate, p.PositionId); err != nil {
		return err
	}
	if err := saveAccount(ctx, tx, a, Now()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// deletePositionRow removes the row and saves the re-margined account in one transaction.
func (h *Handler) deletePositionRow(ctx context.Context, p *model.Position, a *model.Account) error {
	tx, err := h.DB.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`DELETE FROM hst.positions WHERE position_id = $1`, p.PositionId); err != nil {
		return err
	}
	if err := saveAccount(ctx, tx, a, Now()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
