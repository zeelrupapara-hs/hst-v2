package handler

import (
	"context"
	"math"

	"hstcore/internal/book"
	"hstcore/model"
)

// gapPoints is how far a price has to jump before the symbol counts as gapped.
// ponytail: one threshold for every instrument, split it per symbol if a broker needs it
const gapPoints = 100

// NotifyAll records a price and works through the accounts it affects.
func (h *Handler) NotifyAll(t model.Tick) {
	previous, had := h.Quotes.Set(t)

	// a jump beyond the threshold puts the symbol in a gap, which the routing rules can see
	if had {
		h.checkGap(t, previous)
	}

	watching := h.Accounts.Watching(t.Symbol)
	if len(watching) == 0 {
		return
	}

	for _, e := range watching {
		entry := e
		h.Workers.Submit(func(ctx context.Context) {
			h.Notify(ctx, entry, t)
		})
	}
}

// checkGap decides whether the price jumped far enough to count as a gap.
func (h *Handler) checkGap(t, previous model.Tick) {
	sym, ok := h.Settings.Symbol(t.Symbol)
	if !ok || sym.Point <= 0 {
		return
	}

	jump := math.Max(math.Abs(t.Bid-previous.Bid), math.Abs(t.Ask-previous.Ask))

	h.Quotes.SetGap(t.Symbol, Points(jump, sym.Point) >= gapPoints)
}

// Notify brings one account up to date with a new price.
func (h *Handler) Notify(ctx context.Context, e *book.Entry, t model.Tick) {
	e.Lock()

	group := e.Account.Group

	r, ok := h.Settings.For(group, t.Symbol)
	if !ok {
		e.Unlock()
		return
	}

	h.Revalue(e, t.Symbol, t)

	money := h.SettleAccount(e, r.Group.MarginFreeProfit != 0)
	money.Apply(e.Account)

	account := *e.Account
	level := money.MarginLevel
	margin := money.Margin

	// what the price crossed, gathered while the account is locked and acted on after
	hits := h.CookPosition(e, t)
	pendings := h.pendingHits(e, t)

	e.Unlock()

	h.PublishAccount(&account)

	// expiry before anything else.
	h.ExpireOrders(ctx, e, t.Symbol)

	for _, hit := range hits {
		kind := model.RouteSL
		if hit.reason == model.ReasonTP {
			kind = model.RouteTP
		}

		h.CloseAtMarket(ctx, e, hit.position, t, hit.reason, kind)
	}

	for _, hit := range pendings {
		_ = h.CookOrder(ctx, e, hit, t)
	}

	// stop out last: closing a position for a stop loss may have already fixed the level
	if margin > 0 && level > 0 {
		h.checkStopOut(ctx, e, r.Group, t)
	}
}
