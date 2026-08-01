package handler

import (
	"context"
	"strings"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/internal/shardmap"
	"hstcore/model"
	"hstcore/pkg/logger"
)

// Configuration is loaded whole into every pod; accounts only for the shards this pod owns.

// load fills the caches, before the first subscription.
func (h *Handler) load(ctx context.Context) error {
	if err := h.loadSettings(ctx); err != nil {
		return err
	}
	if err := h.loadRules(ctx); err != nil {
		return err
	}
	if err := h.loadCommissions(ctx); err != nil {
		return err
	}

	h.Shards = shardmap.New(h.name, []string{h.name})

	if err := h.loadAccounts(ctx); err != nil {
		return err
	}

	h.Log.Log(logger.TypeSys, logger.CodeOK, "engine loaded",
		"pod", h.name,
		"groups", h.Settings.Groups(),
		"symbols", h.Settings.Symbols(),
		"sessions", h.Sessions.Len(),
		"holidays", h.Holidays.Len(),
		"rules", len(h.rules),
		"shards", len(h.Shards.Mine()),
		"accounts", h.Accounts.Len())

	return nil
}

func (h *Handler) loadSettings(ctx context.Context) error {
	groups := make(map[string]*model.Group, 512)

	rows, err := h.DB.DB.Query(ctx,
		`SELECT group_id, "group", currency, currency_digits,
		        margin_mode, margin_free_mode, margin_so_mode, margin_call, margin_stop_out,
		        margin_flags, margin_free_profit_mode,
		        trade_flags, trade_interest_rate, trade_virtual_credit,
		        limit_orders, limit_positions, limit_positions_volume, limit_symbols
		   FROM hst.groups`)
	if err != nil {
		return err
	}

	for rows.Next() {
		g := &model.Group{}
		if err := rows.Scan(&g.GroupId, &g.Group, &g.Currency, &g.CurrencyDigits,
			&g.MarginMode, &g.MarginFreeMode, &g.MarginSOMode, &g.MarginCall, &g.MarginStopOut,
			&g.MarginFlags, &g.MarginFreeProfit,
			&g.TradeFlags, &g.TradeInterestRate, &g.TradeVirtualCredit,
			&g.LimitOrders, &g.LimitPositions, &g.LimitPositionsValue, &g.LimitSymbols); err != nil {
			rows.Close()
			return err
		}
		groups[g.Group] = g
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	symbols := make(map[string]*model.Symbol, 4096)

	rows, err = h.DB.DB.Query(ctx,
		`SELECT symbol_id, symbol, path, digits, point, calc_mode, trade_mode, exec_mode,
		        fill_flags, expir_flags, contract_size, tick_value, tick_size,
		        spread, spread_diff, stops_level, freeze_level,
		        currency_base, currency_profit, currency_margin,
		        volume_min, volume_max, volume_step, volume_limit,
		        margin_initial, margin_maintenance, margin_hedged, margin_flags,
		        swap_mode, swap_long, swap_short, swap_flags, quotes_timeout, order_flags,
		        swap_rate_sunday, swap_rate_monday, swap_rate_tuesday, swap_rate_wednesday,
		        swap_rate_thursday, swap_rate_friday, swap_rate_saturday, swap_year_day,
		        time_start, time_expiration
		   FROM hst.symbols`)
	if err != nil {
		return err
	}

	for rows.Next() {
		s := &model.Symbol{}
		if err := rows.Scan(&s.SymbolId, &s.Symbol, &s.Path, &s.Digits, &s.Point,
			&s.CalcMode, &s.TradeMode, &s.ExecMode, &s.FillFlags, &s.ExpirFlags,
			&s.ContractSize, &s.TickValue, &s.TickSize,
			&s.Spread, &s.SpreadDiff, &s.StopsLevel, &s.FreezeLevel,
			&s.CurrencyBase, &s.CurrencyProfit, &s.CurrencyMargin,
			&s.VolumeMin, &s.VolumeMax, &s.VolumeStep, &s.VolumeLimit,
			&s.MarginInitial, &s.MarginMaintenance, &s.MarginHedged, &s.MarginFlags,
			&s.SwapMode, &s.SwapLong, &s.SwapShort, &s.SwapFlags, &s.QuotesTime, &s.OrderFlags,
			&s.SwapRate[0], &s.SwapRate[1], &s.SwapRate[2], &s.SwapRate[3],
			&s.SwapRate[4], &s.SwapRate[5], &s.SwapRate[6], &s.SwapYearDay,
			&s.TimeStart, &s.TimeExpiration); err != nil {
			rows.Close()
			return err
		}
		symbols[s.Symbol] = s
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	overrides := make(map[int64][]*model.GroupSymbol, 512)

	// config_index is the group's own ordering and the first match wins
	rows, err = h.DB.DB.Query(ctx,
		`SELECT symbol_id, group_id, path, config_index,
		        trade_mode, exec_mode, fill_flags, expir_flags,
		        spread_diff, stops_level, freeze_level,
		        volume_min, volume_max, volume_step, volume_limit,
		        margin_initial, margin_maintenance, margin_hedged, margin_flags,
		        swap_mode, swap_long, swap_short, swap_flags,
		        swap_rate_sunday, swap_rate_monday, swap_rate_tuesday, swap_rate_wednesday,
		        swap_rate_thursday, swap_rate_friday, swap_rate_saturday, swap_year_day,
		        ie_check_mode, ie_timeout, ie_slip_profit, ie_slip_losing, ie_volume_max, ie_flags,
		        re_timeout, re_flags, order_flags, permissions_flags
		   FROM hst.groups_symbols
		  ORDER BY group_id, config_index`)
	if err != nil {
		return err
	}

	for rows.Next() {
		o := &model.GroupSymbol{}
		if err := rows.Scan(&o.SymbolId, &o.GroupId, &o.Path, &o.ConfigIndex,
			&o.TradeMode, &o.ExecMode, &o.FillFlags, &o.ExpirFlags,
			&o.SpreadDiff, &o.StopsLevel, &o.FreezeLevel,
			&o.VolumeMin, &o.VolumeMax, &o.VolumeStep, &o.VolumeLimit,
			&o.MarginInitial, &o.MarginMaintenance, &o.MarginHedged, &o.MarginFlags,
			&o.SwapMode, &o.SwapLong, &o.SwapShort, &o.SwapFlags,
			&o.SwapRate[0], &o.SwapRate[1], &o.SwapRate[2], &o.SwapRate[3],
			&o.SwapRate[4], &o.SwapRate[5], &o.SwapRate[6], &o.SwapYearDay,
			&o.IECheckMode, &o.IETimeout, &o.IESlipProfit, &o.IESlipLosing, &o.IEVolumeMax, &o.IEFlags,
			&o.RETimeout, &o.REFlags, &o.OrderFlags, &o.PermissionsFlags); err != nil {
			rows.Close()
			return err
		}
		overrides[o.GroupId] = append(overrides[o.GroupId], o)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	h.Settings.Load(groups, symbols, overrides)

	if err := h.loadCalendar(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Handler) loadCalendar(ctx context.Context) error {
	windows := make(map[int64][]settings.Window, 4096)

	rows, err := h.DB.DB.Query(ctx,
		`SELECT symbol_id, type, day, open, close FROM hst.symbols_sessions`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var symbolId int64
		var w settings.Window
		if err := rows.Scan(&symbolId, &w.Type, &w.Day, &w.Open, &w.Close); err != nil {
			rows.Close()
			return err
		}
		windows[symbolId] = append(windows[symbolId], w)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	h.Sessions.Load(windows)

	var days []settings.Holiday

	rows, err = h.DB.DB.Query(ctx,
		`SELECT year, month, day, "from", "to", symbols, mode
		   FROM hst.holidays ORDER BY config_index`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var d settings.Holiday
		var mode int32
		if err := rows.Scan(&d.Year, &d.Month, &d.Day, &d.From, &d.To,
			&d.Symbols, &mode); err != nil {
			rows.Close()
			return err
		}
		d.Enable = mode == 1
		days = append(days, d)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	h.Holidays.Load(days)

	return nil
}

// loadRules reads the routing list in evaluation order, with each rule's conditions and dealers.
func (h *Handler) loadRules(ctx context.Context) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT routing_id, name, mode, request, type, flags, action, action_value, routing_index
		   FROM hst.routing
		  ORDER BY routing_index`)
	if err != nil {
		return err
	}

	var rules []model.RoutingRule
	byId := make(map[int64]int, 32)

	for rows.Next() {
		var r model.RoutingRule
		if err := rows.Scan(&r.RoutingId, &r.Name, &r.Mode, &r.Request, &r.Type,
			&r.Flags, &r.Action, &r.ActionValue, &r.Index); err != nil {
			rows.Close()
			return err
		}
		byId[r.RoutingId] = len(rules)
		rules = append(rules, r)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	rows, err = h.DB.DB.Query(ctx,
		`SELECT condition_id, routing_id, condition, rule, value
		   FROM hst.routing_conds ORDER BY routing_id, condition_id`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var routingId int64
		var c model.RoutingCondition
		if err := rows.Scan(&c.ConditionId, &routingId, &c.Condition, &c.Rule, &c.Value); err != nil {
			rows.Close()
			return err
		}
		if i, ok := byId[routingId]; ok {
			rules[i].Conditions = append(rules[i].Conditions, c)
		}
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	rows, err = h.DB.DB.Query(ctx,
		`SELECT routing_id, login FROM hst.routing_dealers ORDER BY routing_id, login`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var routingId, login int64
		if err := rows.Scan(&routingId, &login); err != nil {
			rows.Close()
			return err
		}
		if i, ok := byId[routingId]; ok {
			rules[i].Dealers = append(rules[i].Dealers, login)
		}
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	h.mu.Lock()
	h.rules = rules
	h.mu.Unlock()

	if !h.canExecute(rules) {
		h.Log.Log(logger.TypeTrade, logger.CodeAtt,
			"no routing rule can execute a trade: every request will be refused until one is added",
			"rules", len(rules))
	}

	return nil
}

func (h *Handler) canExecute(rules []model.RoutingRule) bool {
	for i := range rules {
		if rules[i].Enabled() && model.RouteAction(rules[i].Action).Executes() {
			return true
		}
	}
	return false
}

// loadAccounts reads this pod's accounts with their open orders and positions.
// LoadAccount takes on a single account, for one that appeared after boot.
func (h *Handler) LoadAccount(ctx context.Context, login int64) error {
	return h.readAccounts(ctx, `AND u.login = $1`, login)
}

func (h *Handler) loadAccounts(ctx context.Context) error {
	return h.readAccounts(ctx, "")
}

func (h *Handler) readAccounts(ctx context.Context, and string, args ...any) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT u.login, u."group", u.rights, u.leverage,
		        a.currency_digits, a.balance, a.credit, a.margin, a.margin_free,
		        a.margin_level, a.margin_initial, a.margin_maintenance, a.profit,
		        a.storage, a.floating, a.equity, a.assets, a.liabilities,
		        a.blocked_commission, a.blocked_profit, a.updated_at,
		        COALESCE(g.currency, '')
		   FROM hst.users u
		   JOIN hst.accounts a ON a.login = u.login
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE TRUE `+and, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		a := &model.Account{}
		if err := rows.Scan(&a.Login, &a.Group, &a.Rights, &a.Leverage,
			&a.CurrencyDigits, &a.Balance, &a.Credit, &a.Margin, &a.MarginFree,
			&a.MarginLevel, &a.MarginInitial, &a.MarginMaintenance, &a.Profit,
			&a.Storage, &a.Floating, &a.Equity, &a.Assets, &a.Liabilities,
			&a.Commission, &a.BlockedProfit, &a.UpdatedAt, &a.Currency); err != nil {
			return err
		}

		if !h.Shards.HoldsLogin(a.Login) {
			continue
		}

		h.Accounts.Add(&book.Entry{
			Account:   a,
			Orders:    make(map[int64]*model.Order, 4),
			Positions: make(map[int64]*model.Position, 4),
		})
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	if err := h.loadPositions(ctx, and, args...); err != nil {
		return err
	}

	return h.loadOrders(ctx, and, args...)
}

func (h *Handler) loadPositions(ctx context.Context, and string, args ...any) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT position_id, login, dealer, symbol, action, digits, digits_currency, reason,
		        contract_size, time_create, time_update, price_open, price_current,
		        price_sl, price_tp, volume, volume_ext, profit, storage, rate_profit,
		        rate_margin, expert_id, comment, activation_flags
		   FROM hst.positions
		  WHERE TRUE `+strings.ReplaceAll(and, "u.login", "login"), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := &model.Position{}
		if err := rows.Scan(&p.PositionId, &p.Login, &p.Dealer, &p.Symbol, &p.Action,
			&p.Digits, &p.DigitsCurrency, &p.Reason, &p.ContractSize,
			&p.TimeCreate, &p.TimeUpdate, &p.PriceOpen, &p.PriceCurrent,
			&p.PriceSL, &p.PriceTP, &p.Volume, &p.VolumeExt, &p.Profit, &p.Storage,
			&p.RateProfit, &p.RateMargin, &p.ExpertId, &p.Comment,
			&p.ActivationFlags); err != nil {
			return err
		}

		e, ok := h.Accounts.Get(p.Login)
		if !ok {
			continue
		}

		e.Lock()
		e.Positions[p.PositionId] = p
		e.Unlock()

		h.Accounts.Watch(p.Symbol, e)
	}

	return rows.Err()
}

// loadOrders puts every working order back on its account; history stays on disk.
func (h *Handler) loadOrders(ctx context.Context, and string, args ...any) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT order_id, login, dealer, symbol, digits, digits_currency, contract_size,
		        state, reason, time_setup, time_expiration, time_done, type, type_fill,
		        type_time, price_order, price_trigger, price_current, price_sl, price_tp,
		        volume_initial, volume_current, volume_current_ext, expert_id, position_id,
		        position_by_id, comment, rate_margin,
		        activation_mode, activation_time, activation_price, activation_flags
		   FROM hst.orders
		  WHERE state IN ($1, $2, $3, $4, $5, $6)`,
		int32(model.StateStarted), int32(model.StatePlaced), int32(model.StatePartial),
		int32(model.StateRequestAdd), int32(model.StateRequestModify),
		int32(model.StateRequestCancel))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		o := &model.Order{}
		if err := rows.Scan(&o.OrderId, &o.Login, &o.Dealer, &o.Symbol, &o.Digits,
			&o.DigitsCurrency, &o.ContractSize, &o.State, &o.Reason, &o.TimeSetup,
			&o.TimeExpiration, &o.TimeDone, &o.Type, &o.TypeFill, &o.TypeTime,
			&o.PriceOrder, &o.PriceTrigger, &o.PriceCurrent, &o.PriceSL, &o.PriceTP,
			&o.VolumeInitial, &o.VolumeCurrent, &o.VolumeExt, &o.ExpertId, &o.PositionId,
			&o.PositionById, &o.Comment, &o.RateMargin,
			&o.ActivationMode, &o.ActivationTime, &o.ActivationPrice,
			&o.ActivationFlags); err != nil {
			return err
		}

		e, ok := h.Accounts.Get(o.Login)
		if !ok {
			continue
		}

		e.Lock()
		e.Orders[o.OrderId] = o
		e.Unlock()

		h.Accounts.Watch(o.Symbol, e)
	}

	return rows.Err()
}

func shiftArgs(and string) string { return strings.ReplaceAll(and, "$1", "$2") }

// RefreshAccount reads back what a manager may have changed about an account — its group, its
// leverage, its rights — and works the money out again at the new settings.
func (h *Handler) RefreshAccount(ctx context.Context, e *book.Entry) error {
	e.Lock()

	login := e.Account.Login

	if err := h.DB.DB.QueryRow(ctx,
		`SELECT u."group", u.rights, u.leverage, COALESCE(g.currency, ''), a.currency_digits
		   FROM hst.users u
		   JOIN hst.accounts a ON a.login = u.login
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE u.login = $1`, login).Scan(&e.Account.Group, &e.Account.Rights,
		&e.Account.Leverage, &e.Account.Currency, &e.Account.CurrencyDigits); err != nil {
		e.Unlock()
		return err
	}

	group, _ := h.Settings.Group(e.Account.Group)
	h.SettleAccount(e, group != nil && group.MarginFreeProfit != 0).Apply(e.Account)

	account := *e.Account
	e.Unlock()

	if err := h.SaveAccount(ctx, &account); err != nil {
		return err
	}

	h.PublishWS(model.SubjectAccountSummary(login), "account", &account)

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "account refreshed",
		"login", login, "group", account.Group, "leverage", account.Leverage)

	return nil
}
