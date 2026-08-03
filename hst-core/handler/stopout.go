package handler

import (
	"hstcore/internal/settings"
	wire "hstmodel"

	"context"
	"sort"

	"hstcore/internal/book"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Both levels come from the group.

// checkStopOut looks at where the account stands and closes positions if it must.
func (h *Handler) checkStopOut(ctx context.Context, e *book.Entry, g *model.Group, t model.Tick) {
	e.Lock()

	money := h.CalculateAccountMargins(e)

	// nothing reserved means nothing at risk, unless the group asks us to look at a fully
	// covered book: there the margin is zero while the equity can still go under
	fullyHedged := g.TradeFlags&model.TradeFlagSOFullyHedged != 0 &&
		model.MarginMode(g.MarginMode).Hedging()

	if money.Margin <= 0 && !(fullyHedged && money.Equity < 0) {
		e.Unlock()
		return
	}

	call, stop := g.MarginCall, g.MarginStopOut
	level := money.MarginLevel

	// the levels can be read as money rather than a percentage
	if model.StopOutMode(g.MarginSOMode) == model.StopOutMode_money {
		level = money.Equity
	}

	switch {
	case money.Margin <= 0 && fullyHedged && money.Equity < 0:
		// a covered book with negative equity, which no level would ever catch
	case stop > 0 && level <= stop:
		// fall through and close
	case call > 0 && level <= call:
		login := e.Account.Login
		e.Unlock()

		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "margin call",
			"login", login, "level", level, "call_at", call)

		h.PublishWS(model.SubjectAccountMarginCall(login), wire.EventMarginCall, map[string]any{
			"login": login, "margin_level": level, "call_level": call,
		})

		return
	default:
		e.Unlock()
		return
	}

	worst := h.stopOutOrder(e, g)

	login := e.Account.Login
	e.Unlock()

	h.Log.Log(logger.TypeTrade, logger.CodeAtt, "stop out",
		"login", login, "level", level, "stop_at", stop, "positions", len(worst))

	for _, p := range worst {
		h.logStopOut(ctx, e, p, money, stop)
		h.CloseAtMarket(ctx, e, p, h.tickFor(h.rulesFor(e, p.Symbol), p.Symbol, t), model.OrderReason_so,
			model.RouteFlags_stop_out_position)

		// stop as soon as the account is back above the line
		e.Lock()
		after := h.CalculateAccountMargins(e)
		e.Unlock()

		if after.Margin <= 0 || after.MarginLevel > stop {
			break
		}
	}

	h.CompensateNegativeBalance(ctx, e, g)
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

	sort.Slice(worst, func(i, j int) bool { return worst[i].Profit < worst[j].Profit })

	return worst
}

// CompensateNegativeBalance brings a balance the stop out left short back to zero, which is
// what a broker who does not chase clients for a debt has to do.
func (h *Handler) CompensateNegativeBalance(ctx context.Context, e *book.Entry, g *model.Group) {
	if g.TradeFlags&model.TradeFlagSOCompensation == 0 {
		return
	}

	e.Lock()
	balance, credit, login := e.Account.Balance, e.Account.Credit, e.Account.Login
	e.Unlock()

	if balance >= 0 {
		return
	}

	h.NewBalance(ctx, &model.BalanceRequest{
		Login:         login,
		Action:        model.DealAction_so_compensation,
		Amount:        -balance,
		AllowNegative: true,
		Comment:       "so compensation",
	})

	// the credit that backed the lost position goes with it
	if g.TradeFlags&model.TradeFlagSOCompensationCredit != 0 && credit != 0 {
		h.NewBalance(ctx, &model.BalanceRequest{
			Login:         login,
			Action:        model.DealAction_so_compensation_cr,
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
