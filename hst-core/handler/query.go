package handler

import (
	"encoding/json"
	"sort"

	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// Reading live state out of the engine.
//
// Orders, positions and balances are written down, but what a position is worth right now is
// not: it is worked out from the price on every tick and kept in memory. A reader going to the
// database therefore sees the account as it was when it last traded, not as it stands. These
// answer with what the engine actually holds.

func (h *Handler) QuerySystemEventHandler(msg *natscore.Msg) {
	var q model.QueryRequest
	if err := json.Unmarshal(msg.Data, &q); err != nil {
		h.Log.Log(logger.TypeAPI, logger.CodeWarn, "bad query", "error", err.Error())
		return
	}

	if msg.Reply == "" {
		return
	}

	res := &model.QueryResult{Login: q.Login}

	switch q.What {
	case model.QueryAccount:
		res.Account, res.Found = h.GetAccount(q.Login)
	case model.QueryPositions:
		res.Positions, res.Found = h.GetAllPositions(q.Login)
	case model.QueryOrders:
		res.Orders, res.Found = h.GetAllOrders(q.Login)
	case model.QuerySymbols:
		res.Symbols, res.Found = h.GetAllSymbols(q.Login)
	case model.QueryState:
		res.Account, res.Found = h.GetAccount(q.Login)
		res.Positions, _ = h.GetAllPositions(q.Login)
		res.Orders, _ = h.GetAllOrders(q.Login)
	}

	body, err := json.Marshal(res)
	if err != nil {
		return
	}

	if err := msg.Respond(body); err != nil {
		h.Log.Log(logger.TypeNet, logger.CodeWarn, "could not answer a query",
			"login", q.Login, "error", err.Error())
	}
}

func (h *Handler) GetAccount(login int64) (*model.Account, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	a := *e.Account

	return &a, true
}

func (h *Handler) GetAllPositions(login int64) ([]model.Position, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	out := make([]model.Position, 0, len(e.Positions))
	for _, p := range e.Positions {
		out = append(out, *p)
	}

	return out, true
}

func (h *Handler) GetAllOrders(login int64) ([]model.Order, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	defer e.Unlock()

	out := make([]model.Order, 0, len(e.Orders))
	for _, o := range e.Orders {
		out = append(out, *o)
	}

	return out, true
}

// GetAllSymbols is every instrument the account's group may trade, resolved and priced.
//
// The engine answers rather than the database because the group override is folded onto the
// instrument here, and the ticket has to be shown the same limits the engine will judge the
// order against. Settings.For reports whether the group reaches the symbol at all.
func (h *Handler) GetAllSymbols(login int64) ([]model.SymbolInfo, bool) {
	e, ok := h.Accounts.Get(login)
	if !ok {
		return nil, false
	}

	e.Lock()
	group := e.Account.Group
	e.Unlock()

	names := h.Settings.SymbolNames()
	out := make([]model.SymbolInfo, 0, len(names))

	for _, name := range names {
		r, ok := h.Settings.For(group, name)
		if !ok {
			continue
		}

		out = append(out, h.symbolInfo(r))
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })

	return out, true
}

// symbolInfo flattens one resolved rule set, in the units a terminal reads.
func (h *Handler) symbolInfo(r *settings.Rules) model.SymbolInfo {
	v := model.SymbolInfo{
		Symbol:      r.Symbol.Symbol,
		Path:        r.Symbol.Path,
		Description: r.Symbol.Description,

		Digits:       r.Digits,
		Point:        r.Point,
		ContractSize: r.ContractSize,
		TickValue:    r.TickValue,
		TickSize:     r.TickSize,
		CalcMode:     int32(r.CalcMode),
		TradeMode:    int32(r.TradeMode),
		ExecMode:     int32(r.ExecMode),
		FillFlags:    r.FillFlags,
		ExpirFlags:   r.ExpirFlags,
		OrderFlags:   r.OrderFlags,

		VolumeMin:   model.Lots(r.VolumeMin),
		VolumeMax:   model.Lots(r.VolumeMax),
		VolumeStep:  model.Lots(r.VolumeStep),
		VolumeLimit: model.Lots(r.VolumeLimit),

		StopsLevel:  r.StopsLevel,
		FreezeLevel: r.FreezeLevel,
		SpreadDiff:  r.SpreadDiff,

		CurrencyBase:   r.CurrencyBase,
		CurrencyProfit: r.CurrencyProfit,
		CurrencyMargin: r.CurrencyMargin,

		MarginInitial:     r.MarginInitial,
		MarginMaintenance: r.MarginMaintenance,
		MarginHedged:      r.MarginHedged,

		SwapMode:  r.SwapMode,
		SwapLong:  r.SwapLong,
		SwapShort: r.SwapShort,
	}

	if t, ok := h.Quotes.Get(r.Symbol.Symbol); ok {
		v.Bid, v.Ask, v.Last, v.Time = t.Bid, t.Ask, t.Last, t.Time
		v.Gap, v.HasQuote = t.Gap, true
	}

	return v
}
