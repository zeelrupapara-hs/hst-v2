package handler

import (
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

	money := h.SettleAccount(e, g.MarginFreeProfit != 0)

	// nothing reserved means nothing at risk, whatever the level reads
	if money.Margin <= 0 {
		e.Unlock()
		return
	}

	call, stop := g.MarginCall, g.MarginStopOut
	level := money.MarginLevel

	// the levels can be read as money rather than a percentage
	if model.StopOutMode(g.MarginSOMode) == model.StopOutMoney {
		level = money.Equity
	}

	switch {
	case stop > 0 && level <= stop:
		// fall through and close
	case call > 0 && level <= call:
		login := e.Account.Login
		e.Unlock()

		h.Log.Log(logger.TypeTrade, logger.CodeWarn, "margin call",
			"login", login, "level", level, "call_at", call)

		h.PublishWS(model.SubjectAccountSummary(login), "margin_call", map[string]any{
			"login": login, "margin_level": level, "call_level": call,
		})

		return
	default:
		e.Unlock()
		return
	}

	// biggest loser first: it frees the most margin per position closed
	worst := make([]*model.Position, 0, len(e.Positions))
	for _, p := range e.Positions {
		worst = append(worst, p)
	}
	sort.Slice(worst, func(i, j int) bool { return worst[i].Profit < worst[j].Profit })

	login := e.Account.Login
	e.Unlock()

	h.Log.Log(logger.TypeTrade, logger.CodeAtt, "stop out",
		"login", login, "level", level, "stop_at", stop, "positions", len(worst))

	for _, p := range worst {
		h.logStopOut(ctx, e, p, money, stop)
		h.CloseAtMarket(ctx, e, p, h.tickFor(p.Symbol, t), model.ReasonStopOut,
			model.RouteStopOutPosition)

		// stop as soon as the account is back above the line
		e.Lock()
		after := h.SettleAccount(e, g.MarginFreeProfit != 0)
		e.Unlock()

		// the group can insist the whole book goes, rather than only enough of it
		if g.TradeFlags&model.TradeFlagSOFullyClose != 0 {
			continue
		}

		if after.Margin <= 0 || after.MarginLevel > stop {
			break
		}
	}
}

// tickFor is the current quote for a symbol, falling back to the tick that started this pass.
func (h *Handler) tickFor(symbol string, fallback model.Tick) model.Tick {
	if t, ok := h.Quotes.Get(symbol); ok {
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
