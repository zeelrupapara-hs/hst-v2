package handler

import (
	"hstcore/internal/settings"
	"hstcore/model"
)

// What a trade reserves, what a position is worth, and where the account stands. All pure.

// MarginFor is what one trade of this size reserves, in the deposit currency.
func MarginFor(r *settings.Rules, lots float64, price float64, leverage int32, rate float64) float64 {
	if leverage <= 0 {
		leverage = 1
	}

	var base float64

	switch r.CalcMode {
	case model.CalcForex:
		base = lots * r.ContractSize / float64(leverage)

	case model.CalcForexNoLeverage:
		base = lots * r.ContractSize

	case model.CalcCFD, model.CalcExchStocks:
		base = lots * r.ContractSize * price

	case model.CalcCFDLeverage:
		base = lots * r.ContractSize * price / float64(leverage)

	case model.CalcCFDIndex, model.CalcFutures, model.CalcExchFutures, model.CalcExchFORTS:
		// futures reserve a fixed amount per lot rather than a slice of the notional
		base = lots * r.MarginInitial
		if r.MarginInitial == 0 {
			base = lots * r.ContractSize * price
		}

	default:
		base = lots * r.ContractSize * price
	}

	// an explicit initial margin on the instrument replaces the formula
	if r.MarginInitial > 0 && r.CalcMode != model.CalcFutures {
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
	case model.CalcFutures, model.CalcExchFutures, model.CalcExchFORTS:
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
func Settle(a *model.Account, positions map[int64]*model.Position, freeProfitOnly bool) Money {
	m := Money{
		Balance:    a.Balance,
		Credit:     a.Credit,
		Commission: a.Commission,
	}

	for _, p := range positions {
		m.Floating += p.Profit
		m.Storage += p.Storage
		m.Margin += p.Margin
	}

	m.Equity = m.Balance + m.Credit + m.Floating + m.Storage - m.Commission

	// when the group forbids it, only a floating loss counts against free margin
	usable := m.Floating
	if freeProfitOnly && usable > 0 {
		usable = 0
	}

	m.FreeMargin = m.Balance + m.Credit + usable + m.Storage - m.Commission - m.Margin

	if m.Margin > 0 {
		m.MarginLevel = m.Equity / m.Margin * 100
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
