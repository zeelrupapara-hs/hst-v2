package handler

import (
	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

// What a trade reserves, what a position is worth, and where the account stands. All pure.

// MarginForPosition is what one position reserves, charged at its own side's rate.
func MarginForPosition(r *settings.Rules, p *model.Position, price float64, leverage int32) float64 {
	kind := model.OrderType_buy
	if !p.IsBuy() {
		kind = model.OrderType_sell
	}

	return MarginForTypePlain(r, p.Lots(), price, leverage, p.RateMargin, kind, false)
}

// MarginForType is the same for a named order type, which decides the multiplier applied at the end.
//
// Floating leverage tiers are applied at account level; this path is the plain instrument margin.
func MarginForType(r *settings.Rules, lots float64, price float64, leverage int32, rate float64,
	kind model.OrderType, maintenance bool) float64 {
	return MarginForTypePlain(r, lots, price, leverage, rate, kind, maintenance)
}

// MarginForTypePlain is instrument margin without floating leverage tiers.
func MarginForTypePlain(r *settings.Rules, lots float64, price float64, leverage int32, rate float64,
	kind model.OrderType, maintenance bool) float64 {
	base := marginBasePlain(r, lots, price, leverage, rate)

	multiplier := r.MarginRate.For(kind)
	if maintenance {
		// no maintenance rate means the initial one stands
		if m := r.MarginRateMaintenance.For(kind); m > 0 {
			multiplier = m
		}
	}

	return base * multiplier
}

func marginBasePlain(r *settings.Rules, lots float64, price float64, leverage int32, rate float64) float64 {
	if leverage <= 0 {
		leverage = 1
	}

	divisor := float64(leverage)

	var base float64

	switch r.CalcMode {
	case model.CalcMode_forex:
		base = lots * r.ContractSize / divisor

	case model.CalcMode_forex_no_leverage:
		base = lots * r.ContractSize

	case model.CalcMode_cfd, model.CalcMode_exch_stocks:
		base = lots * r.ContractSize * price

	case model.CalcMode_cfd_leverage:
		base = lots * r.ContractSize * price / divisor

	case model.CalcMode_cfd_index, model.CalcMode_futures, model.CalcMode_exch_futures, model.CalcMode_exch_forts:
		// futures reserve a fixed amount per lot rather than a slice of the notional
		base = lots * r.MarginInitial
		if r.MarginInitial == 0 {
			base = lots * r.ContractSize * price
		}

	default:
		base = lots * r.ContractSize * price
	}

	// an explicit initial margin on the instrument replaces the formula
	if r.MarginInitial > 0 && r.CalcMode != model.CalcMode_futures {
		base = lots * r.MarginInitial
	}

	if rate <= 0 {
		rate = 1
	}

	return base * rate
}

// ProfitFor is what a position is worth right now, in the deposit currency.
func ProfitFor(r *settings.Rules, buy bool, lots, open, current, rate float64) float64 {
	diff := current - open
	if !buy {
		diff = open - current
	}

	var raw float64

	switch r.CalcMode {
	case model.CalcMode_futures, model.CalcMode_exch_futures, model.CalcMode_exch_forts:
		if r.TickSize > 0 {
			raw = diff / r.TickSize * r.TickValue * lots
		} else {
			raw = diff * r.ContractSize * lots
		}

	default:
		raw = diff * r.ContractSize * lots
	}

	if rate <= 0 {
		rate = 1
	}

	return raw * rate
}

// Money is where an account stands once every position has been valued.
type Money struct {
	Balance     float64
	Credit      float64
	Floating    float64
	Storage     float64
	Commission  float64
	Margin      float64
	Equity      float64
	FreeMargin  float64
	MarginLevel float64
}

// Settle adds up an account from its balance and its open positions.
//
// excluded is the floating result of instruments the group told us to leave out of the money
// entirely, and free says how much of what remains counts toward free margin.
// reserved is margin held by something other than an open position, which is what a working
// order does when its type carries a rate.
func Settle(a *model.Account, positions map[int64]*model.Position,
	free model.FreeMarginMode, excluded, reserved float64) Money {
	m := Money{
		Balance:    a.Balance,
		Credit:     a.Credit,
		Commission: a.Commission,
	}

	// virtual credit backs margin but is never the client's money, so it is dropped once the
	// account has nothing open to back
	if a.VirtualCredit > 0 && len(positions) > 0 {
		m.Credit += a.VirtualCredit
	}

	for _, p := range positions {
		m.Floating += p.Profit
		m.Storage += p.Storage
		m.Margin += p.Margin
	}

	m.Margin += reserved

	// an excluded instrument is invisible to the money: not in equity, not in free margin
	counted := m.Floating - excluded

	m.Equity = m.Balance + m.Credit + counted + m.Storage - m.Commission + a.BlockedProfit

	// profit held aside for the day is the client's, but not theirs to trade on yet
	m.FreeMargin = m.Balance + m.Credit + free.Counts(counted) + m.Storage -
		m.Commission - m.Margin

	// the level is read against maintenance margin where the instruments set one, which is what
	// decides a margin call, not the initial reservation
	against := m.Margin
	if a.MarginMaintenance > 0 {
		against = a.MarginMaintenance
	}

	if against > 0 {
		m.MarginLevel = m.Equity / against * 100
	}

	return m
}

// Apply writes a settlement back onto the account.
func (m Money) Apply(a *model.Account) {
	a.Floating = m.Floating
	a.Storage = m.Storage
	a.Margin = m.Margin
	a.Equity = m.Equity
	a.MarginFree = m.FreeMargin
	a.MarginLevel = m.MarginLevel
	a.Profit = m.Floating
}

// NormalisePrice rounds a price to the instrument's digits.
func NormalisePrice(price float64, digits int32) float64 {
	p := 1.0
	for i := int32(0); i < digits; i++ {
		p *= 10
	}

	if p == 0 {
		return price
	}

	// the half added before truncating is what makes this round rather than floor
	if price >= 0 {
		return float64(int64(price*p+0.5)) / p
	}
	return float64(int64(price*p-0.5)) / p
}

// Points turns a price difference into points, the unit stops levels and slippage are set in.
func Points(diff float64, point float64) float64 {
	if point <= 0 {
		return 0
	}
	return diff / point
}

// CalculateAccountSpread is the tick as one group sees it.
//
// A group's spread_diff widens the quote in points around the mid, so two groups on the same
// instrument trade at different prices. The engine keeps one raw book, so the markup is applied
// here, at each price the group is about to be charged, and never written back to the book.
func CalculateAccountSpread(r *settings.Rules, t model.Tick) model.Tick {
	if r == nil || r.Point <= 0 || (r.SpreadDiff == 0 && r.SpreadDiffBalance == 0) {
		return t
	}

	// the difference is split evenly and the balance shifts the split: D=4 B=0 is -2/+2,
	// B=-1 is -3/+1, B=+1 is -1/+3, as the reference defines it
	d := float64(r.SpreadDiff)
	b := float64(r.SpreadDiffBalance)

	t.Bid = NormalisePrice(t.Bid-(d/2-b)*r.Point, r.Digits)
	t.Ask = NormalisePrice(t.Ask+(d/2+b)*r.Point, r.Digits)

	return t
}

// QuoteFor is the price this group trades at: the book's tick with the group's spread applied.
//
// Every path that prices a trade goes through here rather than the book directly, so no route can
// quietly charge the raw feed price.
func (h *Handler) QuoteFor(r *settings.Rules, symbol string) (model.Tick, bool) {
	t, ok := h.Quotes.Get(symbol)
	if !ok {
		return t, false
	}

	return CalculateAccountSpread(r, t), true
}

// The two halves of what a trade costs, kept together because they are easy to miss apart.
//
// Commission is charged once, as the deal is booked. Swap is accrued nightly onto the position and
// only reaches the balance when that position closes, so the money moves in two different places
// at two different times.

// ChargeCommission takes the commission for every deal in a fill off the balance.
func (h *Handler) ChargeCommission(e *book.Entry, f *Fill, r *settings.Rules) {
	for _, d := range f.Deals {
		d.Commission = -h.CommissionFor(d, r)
		e.Account.Balance += d.Commission
	}
}

// SettleSwap moves the swap accrued on a closing position onto the balance.
func (h *Handler) SettleSwap(e *book.Entry, p *model.Position) {
	e.Account.Balance += p.Storage
}
