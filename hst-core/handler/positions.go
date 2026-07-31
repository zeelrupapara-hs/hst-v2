package handler

import (
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

// Turning a filled order into positions and deals.
//
// The whole difference between netting and hedging lives here. Under netting an account holds
// one position per symbol, and an opposite fill shrinks it, closes it, or turns it around.
// Under hedging every fill opens a position of its own and they sit side by side.
//
// Both are described in group_position.htm; which one applies is the group's margin mode.

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

// Execute turns an order into deals and positions on the account. The caller must already hold
// the account's lock.
func (h *Handler) Execute(e *book.Entry, o *model.Order, r *settings.Rules, price float64,
	now int64) *Fill {
	f := &Fill{Price: price, RetCode: model.RetOK}

	if model.MarginMode(r.Group.MarginMode).Hedging() {
		h.openPosition(f, e, o, r, price, now)
		return f
	}

	h.netInto(f, e, o, r, price, now)

	return f
}

// openPosition starts a new position. This is every fill under hedging, and the first fill on a
// symbol under netting.
func (h *Handler) openPosition(f *Fill, e *book.Entry, o *model.Order, r *settings.Rules,
	price float64, now int64) {
	side := int32(0)
	if !o.Kind().Buy() {
		side = 1
	}

	p := &model.Position{
		Login:          o.Login,
		Dealer:         o.Dealer,
		Symbol:         o.Symbol,
		Action:         side,
		Digits:         r.Digits,
		DigitsCurrency: e.Account.CurrencyDigits,
		Reason:         o.Reason,
		ContractSize:   r.ContractSize,
		TimeCreate:     now,
		TimeUpdate:     now,
		PriceOpen:      price,
		PriceCurrent:   price,
		PriceSL:        o.PriceSL,
		PriceTP:        o.PriceTP,
		Volume:         o.VolumeCurrent,
		RateMargin:     o.RateMargin,
		RateProfit:     1,
		ExpertId:       o.ExpertId,
		Comment:        o.Comment,
	}

	p.Margin = MarginFor(r, p.Lots(), price, e.Account.Leverage, p.RateMargin)

	f.Opened = p
	f.Deals = append(f.Deals, h.deal(o, r, e, price, o.VolumeCurrent, 0, model.EntryIn, 0, now))
}

// netInto applies a fill to an account that keeps one position per symbol.
func (h *Handler) netInto(f *Fill, e *book.Entry, o *model.Order, r *settings.Rules,
	price float64, now int64) {
	existing := h.positionOn(e, o.Symbol)

	// nothing open, or the fill is the same way as what is open: open or grow
	if existing == nil {
		h.openPosition(f, e, o, r, price, now)
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
		h.closePosition(f, e, existing, o, r, price, now)
	default:
		h.reversePosition(f, e, existing, o, r, price, now)
	}
}

// growPosition adds to a position and blends the open price. MT5 keeps the weighted average, so
// the position reads as one trade at one price however many fills made it.
func (h *Handler) growPosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	total := p.Volume + o.VolumeCurrent

	p.PriceOpen = (p.PriceOpen*float64(p.Volume) + price*float64(o.VolumeCurrent)) / float64(total)
	p.PriceOpen = NormalisePrice(p.PriceOpen, r.Digits)
	p.Volume = total
	p.TimeUpdate = now
	p.Margin = MarginFor(r, p.Lots(), p.PriceOpen, e.Account.Leverage, p.RateMargin)

	f.Changed = append(f.Changed, p)
	f.Deals = append(f.Deals, h.deal(o, r, e, price, o.VolumeCurrent, 0, model.EntryIn, p.PositionId, now))
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

	d := h.deal(o, r, e, price, closed, closed, model.EntryOut, p.PositionId, now)
	d.Profit = profit
	d.PricePosition = p.PriceOpen
	f.Deals = append(f.Deals, d)
}

// closePosition takes the whole position off.
func (h *Handler) closePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.Buy(), p.Lots(), p.PriceOpen, price, p.RateProfit)

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.deal(o, r, e, price, p.Volume, p.Volume, model.EntryOut, p.PositionId, now)
	d.Profit = profit
	d.PricePosition = p.PriceOpen
	f.Deals = append(f.Deals, d)
}

// reversePosition closes what was open and opens the remainder the other way.
//
// MT5 records this as a single deal marked "in and out" rather than two, so the history reads
// as one action by the client.
func (h *Handler) reversePosition(f *Fill, e *book.Entry, p *model.Position, o *model.Order,
	r *settings.Rules, price float64, now int64) {
	profit := ProfitFor(r, p.Buy(), p.Lots(), p.PriceOpen, price, p.RateProfit)
	remainder := o.VolumeCurrent - p.Volume

	f.Profit += profit
	f.Closed = append(f.Closed, p)

	d := h.deal(o, r, e, price, o.VolumeCurrent, p.Volume, model.EntryInOut, p.PositionId, now)
	d.Profit = profit
	d.PricePosition = p.PriceOpen
	f.Deals = append(f.Deals, d)

	// the leftover volume opens a fresh position the other way round
	side := int32(0)
	if !o.Kind().Buy() {
		side = 1
	}

	np := &model.Position{
		Login:          o.Login,
		Dealer:         o.Dealer,
		Symbol:         o.Symbol,
		Action:         side,
		Digits:         r.Digits,
		DigitsCurrency: e.Account.CurrencyDigits,
		Reason:         o.Reason,
		ContractSize:   r.ContractSize,
		TimeCreate:     now,
		TimeUpdate:     now,
		PriceOpen:      price,
		PriceCurrent:   price,
		PriceSL:        o.PriceSL,
		PriceTP:        o.PriceTP,
		Volume:         remainder,
		RateMargin:     o.RateMargin,
		RateProfit:     1,
		ExpertId:       o.ExpertId,
		Comment:        o.Comment,
	}
	np.Margin = MarginFor(r, np.Lots(), price, e.Account.Leverage, np.RateMargin)

	f.Opened = np
}

// positionOn finds the account's position on a symbol. Under netting there is at most one.
func (h *Handler) positionOn(e *book.Entry, symbol string) *model.Position {
	for _, p := range e.Positions {
		if p.Symbol == symbol {
			return p
		}
	}
	return nil
}

// deal builds the ledger row for a fill.
func (h *Handler) deal(o *model.Order, r *settings.Rules, e *book.Entry, price float64,
	volume, closed int64, entry model.DealEntry, positionId, now int64) *model.Deal {
	action := int32(model.DealBuy)
	if !o.Kind().Buy() {
		action = int32(model.DealSell)
	}

	return &model.Deal{
		Login:          o.Login,
		Dealer:         o.Dealer,
		OrderId:        o.OrderId,
		Action:         action,
		Entry:          int32(entry),
		Digits:         r.Digits,
		DigitsCurrency: e.Account.CurrencyDigits,
		ContractSize:   r.ContractSize,
		Time:           now,
		Symbol:         o.Symbol,
		Price:          price,
		PriceSL:        o.PriceSL,
		PriceTP:        o.PriceTP,
		Volume:         volume,
		VolumeClosed:   closed,
		RateProfit:     1,
		RateMargin:     o.RateMargin,
		ExpertId:       o.ExpertId,
		PositionId:     positionId,
		Comment:        o.Comment,
		TickValue:      r.TickValue,
		TickSize:       r.TickSize,
		Reason:         o.Reason,
	}
}

// Revalue prices every position on a symbol against a new quote and returns what changed.
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
