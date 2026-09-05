package handler

import (
	"hstcore/internal/settings"

	"context"
	"fmt"
	"sort"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Both levels come from the group.

// soLevel is the quantity the group's two levels are compared against: the margin level as a
// percentage, or plain equity when the group reads its levels as money.
func soLevel(g *model.Group, m Money) float64 {
	if model.StopOutMode(g.MarginSOMode) == model.StopOutMode_money {
		return m.Equity
	}
	return m.MarginLevel
}

// checkStopOut looks at where the account stands: a margin call warns once, a stop out deletes
// margined pending orders first, then closes positions biggest loser first, one at a time,
// until the account is back above the line.
func (h *Handler) checkStopOut(ctx context.Context, e *book.Entry, g *model.Group, t model.Tick) {
	if !h.lockHeld(e) {
		return
	}

	if e.StopOutBusy {
		e.Unlock()
		return
	}

	money := h.CalculateAccountMargins(e)

	// nothing reserved means nothing at risk, unless the group asks us to look at a fully
	// covered book: there the margin is zero while the equity can still go under
	fullyHedged := g.TradeFlags&model.TradeFlagSOFullyHedged != 0 &&
		model.MarginMode(g.MarginMode).Hedging()
	hedgedUnder := money.Margin <= 0 && fullyHedged && money.Equity < 0

	if money.Margin <= 0 && !hedgedUnder {
		e.MarginCalled = false
		e.StopOutStarved = false
		e.Unlock()
		return
	}

	call, stop := g.MarginCall, g.MarginStopOut
	level := soLevel(g, money)

	switch {
	case hedgedUnder:
		// a covered book with negative equity, which no level would ever catch
	case stop > 0 && level <= stop:
		// fall through and close
	case call > 0 && level <= call:
		// a warning state, not an action: the account may still trade; said once on the way in
		if e.MarginCalled {
			e.Unlock()
			return
		}
		e.MarginCalled = true
		login := e.Account.Login
		group := e.Account.Group
		e.Unlock()

		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "margin call",
			"login", login, "level", level, "call_at", call)

		warning := map[string]any{
			"login": login, "margin_level": level, "call_level": call,
			"so_mode": g.MarginSOMode,
		}
		h.PublishWS(model.SubjectAccountMarginCall(login), model.EventMarginCall, warning)
		// the managers covering this group hear the same warning
		h.PublishWS(model.SubjectGroupAccounts(group), model.EventMarginCall, warning)

		return
	default:
		// a price wobbling on the line must not ring the bell on every crossing, so the state
		// only clears once the account is safely above it
		if level > call*1.05 {
			e.MarginCalled = false
		}
		e.StopOutStarved = false
		e.Unlock()
		return
	}

	e.StopOutBusy = true
	login := e.Account.Login
	group := e.Account.Group
	e.Unlock()

	defer func() {
		e.Lock()
		e.StopOutBusy = false
		e.Unlock()
	}()

	h.Log.Log(logger.TypeTrade, logger.CodeAtt, "stop out",
		"login", login, "level", level, "stop_at", stop)

	closing := map[string]any{
		"login": login, "margin_level": level, "stop_level": stop,
		"so_mode": g.MarginSOMode,
	}
	h.PublishWS(model.SubjectAccountMarginCall(login), model.EventStopOut, closing)
	h.PublishWS(model.SubjectGroupAccounts(group), model.EventStopOut, closing)

	if h.stopOutPendings(ctx, e, g, stop, hedgedUnder) {
		h.CompensateNegativeBalance(ctx, e, g)
		return
	}

	h.stopOutPositions(ctx, e, g, t, stop, level, hedgedUnder)

	h.CompensateNegativeBalance(ctx, e, g)
}

// soRecovered says whether the account is back above the stop line, on the same measure the
// stop was declared on. A covered book recovers when its equity does.
func soRecovered(g *model.Group, m Money, stop float64, hedgedUnder bool) bool {
	if hedgedUnder {
		return m.Equity >= 0
	}
	return m.Margin <= 0 || soLevel(g, m) > stop
}

// stopOutPendings deletes the working orders that reserve margin, the largest reservation
// first, and reports whether that alone brought the account back.
func (h *Handler) stopOutPendings(ctx context.Context, e *book.Entry, g *model.Group,
	stop float64, hedgedUnder bool) bool {
	type reservedOrder struct {
		order    *model.Order
		reserved float64
	}

	e.Lock()

	candidates := make([]reservedOrder, 0, len(e.Orders))
	for _, o := range e.Orders {
		kind := o.Kind()
		if !kind.IsPending() || !o.State.IsLive() || o.PriceOrder <= 0 {
			continue
		}
		r, ok := h.Settings.For(e.Account.Group, o.Symbol)
		if !ok || r.MarginRate.For(kind) <= 0 {
			continue
		}
		reserved := MarginForType(r, model.Lots(o.VolumeCurrent), o.PriceOrder,
			e.Account.Leverage, o.RateMargin, kind, false)
		if reserved <= 0 {
			// an order with no margin requirement is never deleted for a stop out
			continue
		}
		candidates = append(candidates, reservedOrder{order: o, reserved: reserved})
	}

	e.Unlock()

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].reserved != candidates[j].reserved {
			return candidates[i].reserved > candidates[j].reserved
		}
		return candidates[i].order.OrderId < candidates[j].order.OrderId
	})

	for _, c := range candidates {
		// the rules see the deletion as its own request kind, so a broker can refuse or route it
		r, _ := h.Settings.For(e.Account.Group, c.order.Symbol)
		tick, _ := h.QuoteFor(r, c.order.Symbol)
		decision := h.Route(&Request{
			Kind: model.RouteFlags_stop_out_order, Order: c.order, Entry: e, Rules: r, Tick: tick,
		})
		if decision.Rule != nil && !decision.Executes() &&
			decision.Action != model.RouteAction_cancel_order {
			h.Log.Log(logger.TypeTrade, logger.CodeWarn, "a stop out deletion was refused by a rule",
				"login", c.order.Login, "order", c.order.OrderId, "symbol", c.order.Symbol, "rule", decision.Rule.Name)
			continue
		}

		e.Lock()
		if _, still := e.Orders[c.order.OrderId]; !still || !c.order.State.IsLive() {
			e.Unlock()
			continue
		}
		h.removeOrder(ctx, e, c.order, "stop out")

		e.Lock()
		after := h.CalculateAccountMargins(e)
		e.Unlock()

		if soRecovered(g, after, stop, hedgedUnder) {
			return true
		}
	}

	return false
}

// stopOutPositions closes the account's positions until it recovers, recording the level each
// close happened at.
func (h *Handler) stopOutPositions(ctx context.Context, e *book.Entry, g *model.Group,
	t model.Tick, stop, level float64, hedgedUnder bool) {
	e.Lock()
	worst := h.stopOutOrder(e, g)
	e.Unlock()

	if len(worst) == 0 {
		// everything left is excluded or untradable; said once, or every tick would repeat it
		e.Lock()
		starved := e.StopOutStarved
		e.StopOutStarved = true
		e.Unlock()
		if !starved {
			h.Log.Log(logger.TypeTrade, logger.CodeWarn, "stop out has nothing it may close",
				"login", e.Account.Login, "level", level, "stop_at", stop)
		}
		return
	}

	comment := soComment(g, level)

	for _, p := range worst {
		// the level each close fires at, not the one the pass started with
		e.Lock()
		money := h.CalculateAccountMargins(e)
		e.Unlock()

		h.logStopOut(ctx, e, p, money, stop)
		h.closeAtMarket(ctx, e, p, h.tickFor(h.rulesFor(e, p.Symbol), p.Symbol, t),
			model.OrderReason_so, model.RouteFlags_stop_out_position, comment)

		e.Lock()
		after := h.CalculateAccountMargins(e)
		e.Unlock()

		if soRecovered(g, after, stop, hedgedUnder) {
			break
		}
	}
}

// soComment is the history comment of a stop out close, carrying the level it fired at.
func soComment(g *model.Group, level float64) string {
	if model.StopOutMode(g.MarginSOMode) == model.StopOutMode_money {
		return fmt.Sprintf("[so at %.2f]", level)
	}
	return fmt.Sprintf("[so at %.2f%%]", level)
}

// stopOutOrder is the order positions are taken off in: the biggest loser first, because it
// frees the most margin per close.
//
// Under the first in first out rule only the oldest position on each instrument may be closed,
// so the choice is made among those.
func (h *Handler) stopOutOrder(e *book.Entry, g *model.Group) []*model.Position {
	worst := make([]*model.Position, 0, len(e.Positions))

	oldest := make(map[string]*model.Position, len(e.Positions))
	fifo := g.TradeFlags&model.TradeFlagFIFOClose != 0

	for _, p := range e.Positions {
		// an instrument kept out of the money is not what put the account at risk
		if h.excluded(e.Account.Group, p.Symbol) {
			continue
		}

		if !fifo {
			worst = append(worst, p)
			continue
		}

		if o, seen := oldest[p.Symbol]; !seen || p.TimeCreate < o.TimeCreate {
			oldest[p.Symbol] = p
		}
	}

	for _, p := range oldest {
		worst = append(worst, p)
	}

	sort.SliceStable(worst, func(i, j int) bool {
		if worst[i].Profit != worst[j].Profit {
			return worst[i].Profit < worst[j].Profit
		}
		return worst[i].PositionId < worst[j].PositionId
	})

	return worst
}

// CompensateNegativeBalance brings a balance the stop out left short back to zero, which is
// what a broker who does not chase clients for a debt has to do.
func (h *Handler) CompensateNegativeBalance(ctx context.Context, e *book.Entry, g *model.Group) {
	if g.TradeFlags&model.TradeFlagSOCompensation == 0 {
		return
	}

	e.Lock()
	balance, login, open := e.Account.Balance, e.Account.Login, len(e.Positions)
	e.Unlock()

	// the reference compensates only once the stop out has taken the last position off
	if balance >= 0 || open > 0 {
		return
	}

	h.NewBalance(ctx, &model.BalanceRequest{
		Login:         login,
		Action:        model.DealAction_so_compensation,
		Amount:        -balance,
		AllowNegative: true,
		Comment:       "so compensation",
	})

	h.Log.Log(logger.TypeTrade, logger.CodeOK, "negative balance compensated after stop out",
		"login", login, "amount", -balance)

	if g.TradeFlags&model.TradeFlagSOCompensationCredit == 0 {
		return
	}

	// the credit that backed the lost book goes with it, read after the compensation posted
	e.Lock()
	credit := e.Account.Credit
	e.Unlock()

	if credit != 0 {
		h.NewBalance(ctx, &model.BalanceRequest{
			Login:         login,
			Action:        model.DealAction_so_compensation_credit,
			Amount:        -credit,
			AllowNegative: true,
			Comment:       "so credit compensation",
		})
	}
}

// rulesFor is this account's settings for one instrument, or nil when the group cannot trade it.
func (h *Handler) rulesFor(e *book.Entry, symbol string) *settings.Rules {
	r, ok := h.Settings.For(e.Account.Group, symbol)
	if !ok {
		return nil
	}

	return r
}

// tickFor is the current quote for a symbol at this group's spread, falling back to the tick that
// started this pass.
func (h *Handler) tickFor(r *settings.Rules, symbol string, fallback model.Tick) model.Tick {
	if t, ok := h.QuoteFor(r, symbol); ok {
		return t
	}

	return fallback
}

// logStopOut records why a position was taken off.
func (h *Handler) logStopOut(ctx context.Context, e *book.Entry, p *model.Position,
	money Money, level float64) {
	if _, err := h.DB.DB.Exec(ctx,
		`INSERT INTO hst.stopout_log
		   (login, position_id, symbol, time, equity, margin, margin_level, so_level,
		    volume_closed, profit, comment)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		p.Login, p.PositionId, p.Symbol, Now(), money.Equity, money.Margin,
		money.MarginLevel, level, p.Volume, p.Profit, "stop out"); err != nil {
		h.Log.Log(logger.TypeTrade, logger.CodeErr, "could not record a stop out",
			"login", p.Login, "position", p.PositionId, "error", err.Error())
	}
}
