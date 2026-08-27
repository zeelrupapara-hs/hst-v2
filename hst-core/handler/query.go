package handler

import (
	"encoding/json"
	"sort"
	"time"

	"hstcore/internal/settings"
	"hstcore/model"
	"hstcore/pkg/logger"

	natscore "github.com/nats-io/nats.go"
)

// Live value is only in memory, so a database reader sees the account as it was when it last traded.

func (h *Handler) QuerySystemEventHandler(msg *natscore.Msg) {
	var q model.QueryRequest
	if err := json.Unmarshal(msg.Data, &q); err != nil {
		h.Log.Log(logger.TypeAPI, logger.CodeWarn, "bad query", "error", err.Error())
		return
	}

	if msg.Reply == "" {
		return
	}

	if h.forwarded(msg, "query", q.Login) {
		return
	}

	res := &model.QueryResult{Login: q.Login}

	switch q.What {
	case model.QueryWhat_account:
		res.Account, res.Found = h.GetAccount(q.Login)
	case model.QueryWhat_positions:
		res.Positions, res.Found = h.GetAllPositions(q.Login)
	case model.QueryWhat_orders:
		res.Orders, res.Found = h.GetAllOrders(q.Login)
	case model.QueryWhat_symbols:
		res.Symbols, res.Found = h.GetAllSymbols(q.Login)
	case model.QueryWhat_state:
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

	// An account with no open position is not in the tick path, so nothing has worked its money
	// out since it was loaded: it would answer with the zeroes its row was created with. Settling
	// on read is what makes a flat account report its free margin rather than nothing.
	h.CalculateAccountMargins(e).Apply(e.Account)

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

// GetAllSymbols answers from the engine so the ticket sees the same folded limits an order is judged against.
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

	g, _ := h.Settings.Group(group)

	for _, name := range names {
		r, ok := h.Settings.For(group, name)
		if !ok {
			continue
		}

		out = append(out, h.symbolInfo(r, g))
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })

	return out, true
}

// symbolInfo flattens one resolved rule set, in the units a terminal reads.
func (h *Handler) symbolInfo(r *settings.Rules, g *model.Group) model.SymbolInfo {
	v := model.SymbolInfo{
		Symbol:      r.Symbol.Symbol,
		Path:        r.Symbol.Path,
		Description: r.Symbol.Description,

		International:   r.Symbol.International,
		ISIN:            r.Symbol.ISIN,
		Category:        r.Symbol.Category,
		Exchange:        r.Symbol.Exchange,
		CFI:             r.Symbol.CFI,
		Industry:        r.Symbol.Industry,
		Country:         r.Symbol.Country,
		Basis:           r.Symbol.Basis,
		Source:          r.Symbol.Source,
		Page:            r.Symbol.Page,
		ColorBackground: r.Symbol.ColorBackground,

		Digits:        r.Digits,
		Point:         r.Point,
		ContractSize:  r.ContractSize,
		TickValue:     r.TickValue,
		TickSize:      r.TickSize,
		CalcMode:      int32(r.CalcMode),
		TradeMode:     int32(r.TradeMode),
		ExecMode:      int32(r.ExecMode),
		FillFlags:     r.FillFlags,
		ExpirFlags:    r.ExpirFlags,
		OrderFlags:    r.OrderFlags,
		Sector:        r.Symbol.Sector,
		GtcMode:       r.Symbol.GtcMode,
		TickChartMode: r.Symbol.TickChartMode,

		VolumeMin:   model.Lots(r.VolumeMin),
		VolumeMax:   model.Lots(r.VolumeMax),
		VolumeStep:  model.Lots(r.VolumeStep),
		VolumeLimit: model.Lots(r.VolumeLimit),

		StopsLevel:  r.StopsLevel,
		FreezeLevel: r.FreezeLevel,

		CurrencyBase:   r.CurrencyBase,
		CurrencyProfit: r.CurrencyProfit,
		CurrencyMargin: r.CurrencyMargin,

		MarginInitial:             r.MarginInitial,
		MarginRateInitialBuy:      r.MarginRate.For(model.OrderType_buy),
		MarginRateInitialSell:     r.MarginRate.For(model.OrderType_sell),
		MarginRateMaintenanceBuy:  r.MarginRateMaintenance.For(model.OrderType_buy),
		MarginRateMaintenanceSell: r.MarginRateMaintenance.For(model.OrderType_sell),
		MarginMaintenance:         r.MarginMaintenance,
		MarginHedged:              r.MarginHedged,

		SwapMode:  r.SwapMode,
		SwapLong:  r.SwapLong,
		SwapShort: r.SwapShort,

		SwapRateSunday:    r.SwapRate[0],
		SwapRateMonday:    r.SwapRate[1],
		SwapRateTuesday:   r.SwapRate[2],
		SwapRateWednesday: r.SwapRate[3],
		SwapRateThursday:  r.SwapRate[4],
		SwapRateFriday:    r.SwapRate[5],
		SwapRateSaturday:  r.SwapRate[6],
		SwapYearDay:       r.SwapYearDay,
		SwapFlags:         r.SwapFlags,
	}

	if g != nil {
		if rule := model.MatchLeverageRule(g, r.Symbol.Symbol, r.Symbol.Path); rule != nil {
			v.MarginSpec = marginSpecFromRule(rule)
		}
	}

	if t, ok := h.QuoteFor(r, r.Symbol.Symbol); ok {
		v.Bid, v.Ask, v.Last, v.Time = t.Bid, t.Ask, t.Last, t.Time
		// a price is kept for ever once received, so being priced says nothing about being fed;
		// an instrument whose feed has stopped is reported for what it is rather than as current
		v.Gap, v.HasQuote = t.Gap, h.quoteIsCurrent(t.Time)
		v.Open, v.High, v.Low, v.Close = t.Open, t.High, t.Low, t.Close

		if t.Close != 0 {
			v.Change = t.Last - t.Close
			v.ChangePercent = v.Change / t.Close * 100
		}
	}

	return v
}

// quoteIsCurrent reports whether a tick is recent enough to call the instrument priced.
func (h *Handler) quoteIsCurrent(at int64) bool {
	max := h.Cfg.Engine.QuoteMaxAge
	if max <= 0 || at <= 0 {
		return at > 0
	}

	return time.Since(time.Unix(0, at)) <= max
}

func marginSpecFromRule(rule *model.LeverageRule) *model.SymbolMarginSpec {
	if rule == nil || len(rule.Tiers) == 0 {
		return nil
	}

	spec := &model.SymbolMarginSpec{
		Floating:  true,
		RangeMode: rule.RangeMode,
		RulePath:  rule.Path,
		Tiers:     make([]model.SymbolMarginTier, 0, len(rule.Tiers)),
	}

	var cursor float64
	for i, t := range rule.Tiers {
		tier := model.SymbolMarginTier{
			RangeFrom:             cursor,
			RangeTo:               t.RangeTo,
			MarginRateInitial:     t.MarginRateInitial,
			MarginRateMaintenance: t.MarginRateMaintenance,
		}
		if i == len(rule.Tiers)-1 {
			tier.RangeTo = 0
		}
		spec.Tiers = append(spec.Tiers, tier)
		if t.RangeTo > 0 {
			cursor = t.RangeTo
		}
	}

	return spec
}
