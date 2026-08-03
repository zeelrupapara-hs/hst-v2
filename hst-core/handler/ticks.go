package handler

import (
	wire "hstmodel"

	"context"

	"hstcore/internal/book"
	"hstcore/model"
)

// NotifyAll records a price and works through the accounts it affects.
func (h *Handler) NotifyAll(t model.Tick) {
	h.Quotes.Set(t)

	// hst-quote owns gap detection, it sees the unfiltered stream and the per-symbol settings
	h.Quotes.SetGap(t.Symbol, t.Gap)

	watching := h.Accounts.Watching(t.Symbol)
	if len(watching) == 0 {
		return
	}

	for _, e := range watching {
		entry := e
		h.Workers.Submit(func(ctx context.Context) {
			h.CalculateAccountProfits(ctx, entry, t)
		})
	}
}

// CalculateAccountProfits brings one account up to date with a new price.
func (h *Handler) CalculateAccountProfits(ctx context.Context, e *book.Entry, t model.Tick) {
	e.Lock()

	group := e.Account.Group

	r, ok := h.Settings.For(group, t.Symbol)
	if !ok {
		e.Unlock()
		return
	}

	// this group's price, not the raw book, so a position is valued at what closing it would pay
	t = CalculateAccountSpread(r, t)

	h.CalcPosition(e, t.Symbol, t)

	money := h.CalculateAccountMargins(e)
	money.Apply(e.Account)

	account := *e.Account
	level := money.MarginLevel
	margin := money.Margin

	// what each position on the instrument that moved is now worth, for the client to paint
	profits := make(map[int64]float64, 4)
	for _, p := range e.Positions {
		if p.Symbol == t.Symbol {
			profits[p.PositionId] = p.Profit
		}
	}

	// built while the account is still held, so the line cannot describe a state that never existed
	summary := h.SummaryFor(e, t.Symbol, Now(), profits)

	// what the price crossed, gathered while the account is locked and acted on after
	hits := h.CookPosition(e, t)
	pendings := h.pendingHits(e, t)

	e.Unlock()

	if summary != "" {
		h.PublishText(model.SubjectAccountSummary(account.Login), wire.EventAccountSummary, summary)
	}

	// expiry before anything else.
	h.ExpireOrders(ctx, e, t.Symbol)

	for _, hit := range hits {
		kind := model.RouteSL
		if hit.reason == model.OrderReason_tp {
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
