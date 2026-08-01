package handler

import (
	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
)

type leg struct {
	volume int64
	value  float64
}

func (l *leg) add(volume int64, price float64) {
	l.volume += volume
	l.value += price * float64(volume)
}

func (l *leg) price() float64 {
	if l.volume == 0 {
		return 0
	}
	return l.value / float64(l.volume)
}

// MarginForSymbol is what one account's whole book on one instrument reserves.
//
// Under netting there is a single position and this is the same answer as before. Under
// hedging the legs are aggregated first: same-direction positions merge at their weighted
// average price, and opposite legs cover each other so only the uncovered part is charged in
// full.
func (h *Handler) MarginForSymbol(e *book.Entry, symbol string, r *settings.Rules) float64 {
	var buys, sells leg

	for _, p := range e.Positions {
		if p.Symbol != symbol {
			continue
		}
		if p.Buy() {
			buys.add(p.Volume, p.PriceOpen)
			continue
		}
		sells.add(p.Volume, p.PriceOpen)
	}

	if buys.volume == 0 && sells.volume == 0 {
		return 0
	}

	leverage := e.Account.Leverage
	rate := h.RateMargin(r, e.Account, true)

	if !model.MarginMode(r.Group.MarginMode).Hedging() {
		total := buys.volume + sells.volume
		price := buys.price()
		if sells.volume > buys.volume {
			price = sells.price()
		}
		return MarginFor(r, model.Lots(total), price, leverage, rate)
	}

	// one leg only: nothing is covered
	if buys.volume == 0 || sells.volume == 0 {
		open, price := buys, buys.price()
		if buys.volume == 0 {
			open, price = sells, sells.price()
		}
		return MarginFor(r, model.Lots(open.volume), price, leverage, rate)
	}

	larger, smaller := buys, sells
	if sells.volume > buys.volume {
		larger, smaller = sells, buys
	}

	// with the larger leg option the whole charge is the bigger side and the smaller one is free
	if r.MarginFlags&model.MarginFlagHedgeLargeLeg != 0 {
		return MarginFor(r, model.Lots(larger.volume), larger.price(), leverage, rate)
	}

	uncovered := larger.volume - smaller.volume
	margin := MarginFor(r, model.Lots(uncovered), larger.price(), leverage, rate)

	// the covered part is charged at the instrument's hedged rate, which is often zero
	if r.MarginHedged > 0 {
		margin += model.Lots(smaller.volume) * r.MarginHedged * rate
	}

	return margin
}

// SpreadMargin shares an instrument's margin back over its positions, so a single position
// still reports a number that adds up to the account total.
func (h *Handler) SpreadMargin(e *book.Entry, symbol string, total float64) {
	var volume int64

	for _, p := range e.Positions {
		if p.Symbol == symbol {
			volume += p.Volume
		}
	}

	if volume == 0 {
		return
	}

	for _, p := range e.Positions {
		if p.Symbol == symbol {
			p.Margin = total * float64(p.Volume) / float64(volume)
		}
	}
}

// RemargeAccount recomputes every instrument the account holds. Called after anything that
// changes what is open, because one new position can change the margin on all of them.
func (h *Handler) RemargeAccount(e *book.Entry) {
	seen := make(map[string]bool, len(e.Positions))

	var initial, maintenance float64

	for _, p := range e.Positions {
		if seen[p.Symbol] {
			continue
		}
		seen[p.Symbol] = true

		r, ok := h.Settings.For(e.Account.Group, p.Symbol)
		if !ok {
			continue
		}

		margin := h.MarginForSymbol(e, p.Symbol, r)
		h.SpreadMargin(e, p.Symbol, margin)

		initial += margin
		maintenance += margin * maintenanceRate(r)
	}

	e.Account.MarginInitial = initial
	e.Account.MarginMaintenance = maintenance

	if g, ok := h.Settings.Group(e.Account.Group); ok {
		e.Account.VirtualCredit = g.TradeVirtualCredit
	}
}

// maintenanceRate is what share of the initial margin has to stay covered to avoid a margin
// call. An instrument that names no maintenance margin holds the whole reservation.
func maintenanceRate(r *settings.Rules) float64 {
	if r.MarginMaintenance <= 0 || r.MarginInitial <= 0 {
		return 1
	}

	return r.MarginMaintenance / r.MarginInitial
}

// SettleAccount brings the margin up to date and then adds the account up. Every path that
// needs the money state goes through here, so no caller can settle against stale margin or
// against the wrong free margin rule.
func (h *Handler) SettleAccount(e *book.Entry) Money {
	h.RemargeAccount(e)

	free := model.FreeMarginUsePL
	if g, ok := h.Settings.Group(e.Account.Group); ok {
		free = model.FreeMarginMode(g.MarginFreeMode)
	}

	return Settle(e.Account, e.Positions, free, h.excludedProfit(e))
}

// excludedProfit is the floating result of the instruments the group keeps out of the money.
func (h *Handler) excludedProfit(e *book.Entry) float64 {
	var out float64

	for _, p := range e.Positions {
		r, ok := h.Settings.For(e.Account.Group, p.Symbol)
		if ok && r.MarginFlags&model.MarginFlagExcludePL != 0 {
			out += p.Profit
		}
	}

	return out
}

// excluded reports whether an instrument is kept out of the money, and so out of stop out too.
func (h *Handler) excluded(group, symbol string) bool {
	r, ok := h.Settings.For(group, symbol)

	return ok && r.MarginFlags&model.MarginFlagExcludePL != 0
}
