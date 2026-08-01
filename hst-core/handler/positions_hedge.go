package handler

import (
	"context"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"
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
func (h *Handler) MarginForSymbol(e *book.Entry, symbol string, r *settings.Rules, maintenance bool) float64 {
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
		price, kind := buys.price(), model.OrderBuy
		if sells.volume > buys.volume {
			price, kind = sells.price(), model.OrderSell
		}
		return MarginForType(r, model.Lots(total), price, leverage, rate, kind, maintenance)
	}

	// one leg only: nothing is covered
	if buys.volume == 0 || sells.volume == 0 {
		open, price, kind := buys, buys.price(), model.OrderBuy
		if buys.volume == 0 {
			open, price, kind = sells, sells.price(), model.OrderSell
		}
		return MarginForType(r, model.Lots(open.volume), price, leverage, rate, kind, maintenance)
	}

	larger, smaller, kind := buys, sells, model.OrderBuy
	if sells.volume > buys.volume {
		larger, smaller, kind = sells, buys, model.OrderSell
	}

	// with the larger leg option the whole charge is the bigger side and the smaller one is free
	if r.MarginFlags&model.MarginFlagHedgeLargeLeg != 0 {
		return MarginForType(r, model.Lots(larger.volume), larger.price(), leverage, rate, kind, maintenance)
	}

	uncovered := larger.volume - smaller.volume
	margin := MarginForType(r, model.Lots(uncovered), larger.price(), leverage, rate, kind, maintenance)

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
func (h *Handler) RemargeAccount(e *book.Entry) float64 {
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

		margin := h.MarginForSymbol(e, p.Symbol, r, false)
		h.SpreadMargin(e, p.Symbol, margin)

		initial += margin
		maintenance += h.MarginForSymbol(e, p.Symbol, r, true) * maintenanceRate(r)
	}

	// a pending order reserves margin of its own where the instrument gives its type a rate
	pending, pendingMaintenance := h.pendingMargin(e)

	e.Account.MarginInitial = initial + pending
	e.Account.MarginMaintenance = maintenance + pendingMaintenance

	if g, ok := h.Settings.Group(e.Account.Group); ok {
		e.Account.VirtualCredit = g.TradeVirtualCredit
	}

	return pending
}

// pendingMargin is what the working orders reserve. An order type whose rate is zero reserves
// nothing, which is what leaving the rate alone means.
func (h *Handler) pendingMargin(e *book.Entry) (initial, maintenance float64) {
	for _, o := range e.Orders {
		kind := o.Kind()
		if !kind.Pending() || !model.OrderState(o.State).Live() {
			continue
		}

		r, ok := h.Settings.For(e.Account.Group, o.Symbol)
		if !ok || r.MarginRate.For(kind) <= 0 {
			continue
		}

		price := o.PriceOrder
		if price <= 0 {
			continue
		}

		lots := model.Lots(o.VolumeCurrent)
		initial += MarginForType(r, lots, price, e.Account.Leverage, o.RateMargin, kind, false)
		maintenance += MarginForType(r, lots, price, e.Account.Leverage, o.RateMargin, kind, true)
	}

	return initial, maintenance
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
	reserved := h.RemargeAccount(e)

	free := model.FreeMarginUsePL
	if g, ok := h.Settings.Group(e.Account.Group); ok {
		free = model.FreeMarginMode(g.MarginFreeMode)
	}

	return Settle(e.Account, e.Positions, free, h.excludedProfit(e), reserved)
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

// SettleAndPublish works the account out again and tells the terminal, for the paths that change
// what is reserved without writing a deal: a working order placed, cancelled or expired.
func (h *Handler) SettleAndPublish(ctx context.Context, e *book.Entry) {
	e.Lock()
	before := e.Account.Margin
	h.SettleAccount(e).Apply(e.Account)
	account := *e.Account
	e.Unlock()

	if account.Margin == before {
		return
	}

	if err := h.SaveAccount(ctx, &account); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not save the reserved margin",
			"login", account.Login, "error", err.Error())
	}

	h.PublishAccount(&account, nil)
}
