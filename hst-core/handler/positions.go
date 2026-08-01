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
	reason   model.Reason
}

// Execute turns an order into deals and positions on the account.
func (h *Handler) Execute(e *book.Entry, o *model.Order, r *settings.Rules, price float64,
	now int64) *Fill {
	f := &Fill{Price: price, RetCode: model.RetOK}

	// an order naming a ticket takes that ticket off, whatever the margin mode.
	if p := h.positionById(e, o.PositionId); p != nil {
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
	side := int32(0)
	if !o.Kind().Buy() {
		side = 1
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
		RateProfit:      1,
		ExpertId:        o.ExpertId,
		Comment:         o.Comment,
		ActivationFlags: o.ActivationFlags,
	}

	p.Margin = MarginFor(r, p.Lots(), price, e.Account.Leverage, p.RateMargin)

	f.Opened = p
	f.Deals = append(f.Deals, h.MakeDealIn(o, r, e, price, o.VolumeCurrent, 0, now))
}

// UpdatePosition moves a position's stop loss and take profit, and nothing else.
func (h *Handler) UpdatePosition(ctx context.Context, req *model.TradeRequest) *model.TradeResult {
	res := &model.TradeResult{RequestId: req.RequestId, Login: req.Login}

	e, ok := h.Accounts.Get(req.Login)
	if !ok {
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

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

	tick, ok := h.Quotes.Get(p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	if code := h.ValidatePosition(e, p, req, r, tick); !code.OK() {
		e.Unlock()
		return h.refuse(res, code, "")
	}

	// the rules see a level change as its own kind of request, and can strip a level off it
	o := h.closingOrder(p, r, model.Reason(req.Reason), p.Volume, "")
	o.PriceSL, o.PriceTP = req.PriceSL, req.PriceTP

	decision := h.Route(&Request{
		Kind: model.RouteSLTP, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.StateRequestModify)
	}

	p.PriceSL, p.PriceTP = o.PriceSL, o.PriceTP
	p.TimeUpdate = Now()

	saved := *p
	e.Unlock()

	if err := h.SavePositionAndPublish(ctx, &saved); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not modify a position",
			"login", saved.Login, "position", saved.PositionId, "error", err.Error())
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
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

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

	tick, ok := h.Quotes.Get(p.Symbol)
	if !ok {
		e.Unlock()
		return h.refuse(res, model.RetTradeNoQuotes, "")
	}

	o := h.closingOrder(p, r, model.Reason(req.Reason), req.Volume, req.Comment)
	o.Dealer = req.Dealer
	o.PriceOrder = req.Price

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
		return h.refuseByRule(res, decision, e, req, o, model.StateRequestModify)
	}

	price := o.PriceOrder
	if decision.AtMarket() || price <= 0 {
		price = tick.ClosePrice(p.Buy())
	}
	price = NormalisePrice(price, r.Digits)

	volume := o.VolumeCurrent
	fill := h.Execute(e, o, r, price, Now())

	h.settle(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close",
			"login", req.Login, "position", req.PositionId, "error", err.Error())
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
		return h.refuse(res, model.RetTradeWrongShard, "")
	}

	e.Lock()

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

	if !model.MarginMode(r.Group.MarginMode).Hedging() {
		e.Unlock()
		return h.refuse(res, model.RetTradeHedgeProhibited, "")
	}
	if p.Symbol != by.Symbol {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "the two positions are on different symbols")
	}
	if p.Buy() == by.Buy() {
		e.Unlock()
		return h.refuse(res, model.RetInvalidData, "the two positions are on the same side")
	}

	tick, ok := h.Quotes.Get(p.Symbol)
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
		State:          int32(model.StateStarted),
		Reason:         req.Reason,
		TimeSetup:      now,
		Type:           int32(model.OrderCloseBy),
		VolumeInitial:  volume,
		VolumeCurrent:  volume,
		PositionById:   by.PositionId,
		RateMargin:     p.RateMargin,
		Comment:        req.Comment,
	}

	decision := h.Route(&Request{
		Kind: model.RouteCloseBy, Order: o, Entry: e, Rules: r, Tick: tick,
		Gapped: h.Quotes.Gapped(p.Symbol),
	})

	if !decision.Executes() {
		e.Unlock()
		return h.refuseByRule(res, decision, e, req, o, model.StateRequestModify)
	}

	f := &Fill{Price: by.PriceOpen, RetCode: model.RetOK}
	h.closeAgainst(f, e, p, by, o, r, volume, now)

	// the ticket goes on after the fill, so Execute is not asked to close it a second time
	o.PositionId = p.PositionId

	h.settle(e, o, f, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, f, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close by",
			"login", req.Login, "position", req.PositionId, "error", err.Error())
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
	t model.Tick, reason model.Reason, kind model.RouteFlags) {
	e.Lock()

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

	o := h.closingOrder(p, r, reason, p.Volume, closeComment(reason))

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

	price := NormalisePrice(t.ClosePrice(p.Buy()), r.Digits)
	fill := h.Execute(e, o, r, price, Now())

	h.settle(e, o, fill, r)

	account := *e.Account
	e.Unlock()

	if err := h.SaveOrderAndPublish(ctx, e, o, fill, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a close",
			"login", p.Login, "position", p.PositionId, "error", err.Error())
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
		price := t.ClosePrice(p.Buy())

		sl := p.PriceSL > 0 && p.ActivationFlags&model.ActivationFlagNoSL == 0
		tp := p.PriceTP > 0 && p.ActivationFlags&model.ActivationFlagNoTP == 0

		switch {
		case sl && p.Buy() && price <= p.PriceSL:
			hits = append(hits, trigger{p, model.ReasonSL})
		case sl && !p.Buy() && price >= p.PriceSL:
			hits = append(hits, trigger{p, model.ReasonSL})
		case tp && p.Buy() && price >= p.PriceTP:
			hits = append(hits, trigger{p, model.ReasonTP})
		case tp && !p.Buy() && price <= p.PriceTP:
			hits = append(hits, trigger{p, model.ReasonTP})
		}
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

	if existing.Buy() == o.Kind().Buy() {
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
	p.Margin = MarginFor(r, p.Lots(), p.PriceOpen, e.Account.Leverage, p.RateMargin)

	f.Changed = append(f.Changed, p)
	f.Deals = append(f.Deals, h.MakeDealIn(o, r, e, price, o.VolumeCurrent, p.PositionId, now))
}

// reducePosition closes part of a position and books the profit on the part that went.
func (h *Handler) reducePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	closed := o.VolumeCurrent
	profit := ProfitFor(r, p.Buy(), model.Lots(closed), p.PriceOpen, price, p.RateProfit)

	p.Volume -= closed
	p.TimeUpdate = now
	p.Margin = MarginFor(r, p.Lots(), p.PriceOpen, e.Account.Leverage, p.RateMargin)

	f.Profit += profit
	f.Changed = append(f.Changed, p)

	d := h.MakeDealOut(o, r, e, p, price, closed, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)
}

// closeInto takes the whole position off.
func (h *Handler) closeInto(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.Buy(), p.Lots(), p.PriceOpen, price, p.RateProfit)

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.MakeDealOut(o, r, e, p, price, p.Volume, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)
}

// reversePosition closes what was open and opens the remainder the other way.
func (h *Handler) reversePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.Buy(), p.Lots(), p.PriceOpen, price, p.RateProfit)
	remainder := o.VolumeCurrent - p.Volume

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.MakeDealOut(o, r, e, p, price, p.Volume, now)
	d.Profit = profit
	f.Deals = append(f.Deals, d)

	// the leftover volume opens a fresh position the other way round
	side := int32(0)
	if !o.Kind().Buy() {
		side = 1
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
		RateProfit:      1,
		ExpertId:        o.ExpertId,
		Comment:         o.Comment,
		ActivationFlags: o.ActivationFlags,
	}
	np.Margin = MarginFor(r, np.Lots(), price, e.Account.Leverage, np.RateMargin)

	f.Opened = np
}

// closeAgainst offsets two opposite positions.
func (h *Handler) closeAgainst(f *Fill, e *book.Entry, p, by *model.Position, o *model.Order,
	r *settings.Rules, volume, now int64) {
	profit := ProfitFor(r, p.Buy(), model.Lots(volume), p.PriceOpen, by.PriceOpen, p.RateProfit)
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

	p.Volume -= volume
	p.TimeUpdate = now
	p.Margin = MarginFor(r, p.Lots(), p.PriceOpen, e.Account.Leverage, p.RateMargin)

	f.Changed = append(f.Changed, p)
}

// SavePositionAndPublish writes a level change and announces it.
func (h *Handler) SavePositionAndPublish(ctx context.Context, p *model.Position) error {
	if _, err := h.DB.DB.Exec(ctx,
		`UPDATE hst.positions
		    SET price_sl = $1, price_tp = $2, activation_flags = $3, time_update = $4,
		        date_modified = $4
		  WHERE position_id = $5`,
		p.PriceSL, p.PriceTP, p.ActivationFlags, p.TimeUpdate, p.PositionId); err != nil {
		return err
	}

	h.PublishWS(model.SubjectAccountPositions(p.Login), "position_changed", p)

	return nil
}

// SavePositionAndPublishAsync hands the write to the worker pool.
func (h *Handler) SavePositionAndPublishAsync(p *model.Position) {
	h.Workers.Submit(func(ctx context.Context) {
		if err := h.SavePositionAndPublish(ctx, p); err != nil {
			h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save a position",
				"login", p.Login, "position", p.PositionId, "error", err.Error())
		}
	})
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
func (h *Handler) positionById(e *book.Entry, id int64) *model.Position {
	if id <= 0 {
		return nil
	}
	return e.Positions[id]
}

// closingOrder is the order that takes volume off a position.
func (h *Handler) closingOrder(p *model.Position, r *settings.Rules, reason model.Reason,
	volume int64, comment string) *model.Order {
	side := int32(model.OrderSell)
	if !p.Buy() {
		side = int32(model.OrderBuy)
	}

	if volume <= 0 || volume > p.Volume {
		volume = p.Volume
	}

	return &model.Order{
		Login:         p.Login,
		Symbol:        p.Symbol,
		Digits:        r.Digits,
		ContractSize:  r.ContractSize,
		State:         int32(model.StateStarted),
		Reason:        int32(reason),
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
func positionReason(reason int32) int32 {
	switch model.Reason(reason) {
	case model.ReasonSL, model.ReasonTP, model.ReasonStopOut:
		return int32(model.ReasonClient)
	}
	return reason
}

// closeComment is what shows against the deal in the client's history.
func closeComment(reason model.Reason) string {
	switch reason {
	case model.ReasonSL:
		return "[sl]"
	case model.ReasonTP:
		return "[tp]"
	case model.ReasonStopOut:
		return "[so]"
	}
	return ""
}

// Revalue prices every position on a symbol against a new quote.
func (h *Handler) Revalue(e *book.Entry, symbol string, t model.Tick) {
	for _, p := range e.Positions {
		if p.Symbol != symbol {
			continue
		}

		r, ok := h.Settings.For(e.Account.Group, symbol)
		if !ok {
			continue
		}

		p.PriceCurrent = t.ClosePrice(p.Buy())
		p.Profit = ProfitFor(r, p.Buy(), p.Lots(), p.PriceOpen, p.PriceCurrent, p.RateProfit)
		p.Margin = MarginFor(r, p.Lots(), p.PriceOpen, e.Account.Leverage, p.RateMargin)
	}
}

// Now is the engine's clock, in epoch nanoseconds.
func Now() int64 { return time.Now().UnixNano() }
