package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"hstcore/internal/book"
	"hstcore/internal/settings"
	"hstcore/internal/shardmap"
	"hstcore/model"
	"hstcore/pkg/logger"

	"github.com/jackc/pgx/v5"
)

// Configuration is loaded whole into every pod; accounts only for the shards this pod owns.

// load fills the caches, before the first subscription.
// StartMarket fills memory from the database. Everything the engine prices against lives in
// memory, so on a restart it is read back in the order it depends on: configuration first, then
// the accounts, then what those accounts already had open when the pod went down.
func (h *Handler) StartMarket(ctx context.Context) error {
	// configuration, replicated in every pod
	if err := h.LoadSettings(ctx); err != nil {
		return err
	}
	if err := h.LoadRoutingRules(ctx); err != nil {
		return err
	}
	if err := h.LoadSystemConfig(ctx); err != nil {
		return err
	}
	if err := h.LoadCommission(ctx); err != nil {
		return err
	}
	if err := h.LoadManagerGroups(ctx); err != nil {
		return err
	}

	// the last price of every instrument, so a restart is not blind until each one next prints
	if err := h.LoadMarketData(ctx); err != nil {
		return err
	}

	// which accounts belong to this pod
	h.Shards = shardmap.New(h.name, []string{h.name})

	// accounts, and what they had open: this is the case of a crash
	if err := h.LoadAccount(ctx); err != nil {
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
		"accounts", h.Accounts.Len(),
		"positions", h.Accounts.Positions(),
		"orders", h.Accounts.Orders())

	return nil
}

func (h *Handler) LoadSettings(ctx context.Context) error {
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
		`SELECT symbol_id, symbol, path, description, digits, point, calc_mode, trade_mode, exec_mode,
		        fill_flags, expir_flags, contract_size, tick_value, tick_size,
		        spread, spread_diff, stops_level, freeze_level,
		        currency_base, currency_profit, currency_margin,
		        volume_min, volume_max, volume_step, volume_limit,
		        volume_min_ext, volume_max_ext, volume_step_ext, volume_limit_ext,
		        margin_initial, margin_maintenance, margin_hedged, margin_flags,
		        margin_initial_buy, margin_initial_sell,
		        margin_initial_buy_limit, margin_initial_sell_limit,
		        margin_initial_buy_stop, margin_initial_sell_stop,
		        margin_initial_buy_stop_limit, margin_initial_sell_stop_limit,
		        margin_maintenance_buy, margin_maintenance_sell,
		        margin_maintenance_buy_limit, margin_maintenance_sell_limit,
		        margin_maintenance_buy_stop, margin_maintenance_sell_stop,
		        margin_maintenance_buy_stop_limit, margin_maintenance_sell_stop_limit,
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
		if err := rows.Scan(&s.SymbolId, &s.Symbol, &s.Path, &s.Description, &s.Digits, &s.Point,
			&s.CalcMode, &s.TradeMode, &s.ExecMode, &s.FillFlags, &s.ExpirFlags,
			&s.ContractSize, &s.TickValue, &s.TickSize,
			&s.Spread, &s.SpreadDiff, &s.StopsLevel, &s.FreezeLevel,
			&s.CurrencyBase, &s.CurrencyProfit, &s.CurrencyMargin,
			&s.VolumeMin, &s.VolumeMax, &s.VolumeStep, &s.VolumeLimit,
			&s.VolumeMinExt, &s.VolumeMaxExt, &s.VolumeStepExt, &s.VolumeLimitExt,
			&s.MarginInitial, &s.MarginMaintenance, &s.MarginHedged, &s.MarginFlags,
			&s.MarginRateInitial[0], &s.MarginRateInitial[1], &s.MarginRateInitial[2], &s.MarginRateInitial[3], &s.MarginRateInitial[4], &s.MarginRateInitial[5], &s.MarginRateInitial[6], &s.MarginRateInitial[7],
			&s.MarginRateMaintenance[0], &s.MarginRateMaintenance[1], &s.MarginRateMaintenance[2], &s.MarginRateMaintenance[3], &s.MarginRateMaintenance[4], &s.MarginRateMaintenance[5], &s.MarginRateMaintenance[6], &s.MarginRateMaintenance[7],
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
		        volume_min_ext, volume_max_ext, volume_step_ext, volume_limit_ext,
		        margin_initial, margin_maintenance, margin_hedged, margin_flags,
		        margin_initial_buy, margin_initial_sell,
		        margin_initial_buy_limit, margin_initial_sell_limit,
		        margin_initial_buy_stop, margin_initial_sell_stop,
		        margin_initial_buy_stop_limit, margin_initial_sell_stop_limit,
		        margin_maintenance_buy, margin_maintenance_sell,
		        margin_maintenance_buy_limit, margin_maintenance_sell_limit,
		        margin_maintenance_buy_stop, margin_maintenance_sell_stop,
		        margin_maintenance_buy_stop_limit, margin_maintenance_sell_stop_limit,
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
			&o.VolumeMinExt, &o.VolumeMaxExt, &o.VolumeStepExt, &o.VolumeLimitExt,
			&o.MarginInitial, &o.MarginMaintenance, &o.MarginHedged, &o.MarginFlags,
			&o.MarginRateInitial[0], &o.MarginRateInitial[1], &o.MarginRateInitial[2], &o.MarginRateInitial[3], &o.MarginRateInitial[4], &o.MarginRateInitial[5], &o.MarginRateInitial[6], &o.MarginRateInitial[7],
			&o.MarginRateMaintenance[0], &o.MarginRateMaintenance[1], &o.MarginRateMaintenance[2], &o.MarginRateMaintenance[3], &o.MarginRateMaintenance[4], &o.MarginRateMaintenance[5], &o.MarginRateMaintenance[6], &o.MarginRateMaintenance[7],
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

	if err := h.LoadCalendar(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Handler) LoadCalendar(ctx context.Context) error {
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
func (h *Handler) LoadRoutingRules(ctx context.Context) error {
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
		`SELECT routing_id, login FROM hst.routing_dealers ORDER BY routing_id, dealer_index`)
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
func (h *Handler) LoadAccountById(ctx context.Context, login int64) error {
	return h.readAccounts(ctx, nil, `AND u.login = $1`, login)
}

func (h *Handler) LoadAccount(ctx context.Context) error {
	return h.readAccounts(ctx, nil, "")
}

// LoadAccountsIn reads only the accounts falling in the shards named, for a pod taking them on.
func (h *Handler) LoadAccountsIn(ctx context.Context, shards map[uint32]bool) error {
	return h.readAccounts(ctx, shards, "")
}

func (h *Handler) readAccounts(ctx context.Context, shards map[uint32]bool, and string, args ...any) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT u.login, u."group", u.rights, u.leverage,
		        COALESCE(g.currency_digits, 2), a.balance, a.credit, a.margin, a.margin_free,
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

		// a rebalance asks for the shards just gained, so the rest are already resident
		if shards != nil && !shards[shardmap.ShardOf(a.Login)] {
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

	if err := h.LoadPositions(ctx, and, args...); err != nil {
		return err
	}

	return h.LoadPendingOrders(ctx, and, args...)
}

func (h *Handler) LoadPositions(ctx context.Context, and string, args ...any) error {
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
		var legacyVolume int64

		if err := rows.Scan(&p.PositionId, &p.Login, &p.Dealer, &p.Symbol, &p.Action,
			&p.Digits, &p.DigitsCurrency, &p.Reason, &p.ContractSize,
			&p.TimeCreate, &p.TimeUpdate, &p.PriceOpen, &p.PriceCurrent,
			&p.PriceSL, &p.PriceTP, &legacyVolume, &p.Volume, &p.Profit, &p.Storage,
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
func (h *Handler) LoadPendingOrders(ctx context.Context, and string, args ...any) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT order_id, login, dealer, symbol, digits, digits_currency, contract_size,
		        state, reason, time_setup, time_expiration, time_done, type, type_fill,
		        type_time, price_order, price_trigger, price_current, price_sl, price_tp,
		        volume_initial, volume_current, volume_current_ext, expert_id, position_id,
		        position_by_id, comment, rate_margin,
		        activation_mode, activation_time, activation_price, activation_flags,
		        routing_id
		   FROM hst.orders
		  WHERE state IN ($1, $2, $3, $4, $5, $6)`,
		model.OrderState_started, model.OrderState_placed, model.OrderState_partial,
		model.OrderState_request_add, model.OrderState_request_modify,
		model.OrderState_request_cancel)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		o := &model.Order{}

		var legacyInitial, legacyCurrent int64

		if err := rows.Scan(&o.OrderId, &o.Login, &o.Dealer, &o.Symbol, &o.Digits,
			&o.DigitsCurrency, &o.ContractSize, &o.State, &o.Reason, &o.TimeSetup,
			&o.TimeExpiration, &o.TimeDone, &o.Type, &o.TypeFill, &o.TypeTime,
			&o.PriceOrder, &o.PriceTrigger, &o.PriceCurrent, &o.PriceSL, &o.PriceTP,
			&legacyInitial, &legacyCurrent, &o.VolumeCurrent, &o.ExpertId, &o.PositionId,
			&o.PositionById, &o.Comment, &o.RateMargin,
			&o.ActivationMode, &o.ActivationTime, &o.ActivationPrice,
			&o.ActivationFlags, &o.RoutingId); err != nil {
			return err
		}

		o.VolumeInitial = model.FromLegacy(legacyInitial)
		o.VolumeCurrent = model.Extended(legacyCurrent, o.VolumeCurrent)

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

// RefreshAccount reads back what a manager may have changed about an account — its group, its
// leverage, its rights — and works the money out again at the new settings.
func (h *Handler) RefreshAccount(ctx context.Context, e *book.Entry) error {
	e.Lock()

	login := e.Account.Login

	if err := h.DB.DB.QueryRow(ctx,
		`SELECT u."group", u.rights, u.leverage,
		        COALESCE(g.currency, ''), COALESCE(g.currency_digits, 2)
		   FROM hst.users u
		   JOIN hst.accounts a ON a.login = u.login
		   LEFT JOIN hst.groups g ON g."group" = u."group"
		  WHERE u.login = $1`, login).Scan(&e.Account.Group, &e.Account.Rights,
		&e.Account.Leverage, &e.Account.Currency, &e.Account.CurrencyDigits); err != nil {
		e.Unlock()
		return err
	}

	h.CalculateAccountMargins(e).Apply(e.Account)

	account := *e.Account
	e.Unlock()

	if err := h.SaveAccount(ctx, &account); err != nil {
		return err
	}

	h.PublishAccount(&account, nil)

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "account refreshed",
		"login", login, "group", account.Group, "leverage", account.Leverage)

	return nil
}

// loadManagerGroups reads which client groups each dealer services.
//
// A dealer can only work an account they are allowed to see, so a request from a group outside
// their masks is not theirs to answer even when a rule names them.
func (h *Handler) LoadManagerGroups(ctx context.Context) error {
	rows, err := h.DB.DB.Query(ctx,
		`SELECT login, COALESCE(groups, '{}') FROM hst.managers`)
	if err != nil {
		return err
	}
	defer rows.Close()

	masks := make(map[int64][]string, 64)

	for rows.Next() {
		var login int64
		var groups []string

		if err := rows.Scan(&login, &groups); err != nil {
			return err
		}

		masks[login] = cleanMasks(groups)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	h.mu.Lock()
	h.managerGroups = masks
	h.mu.Unlock()

	return nil
}

func cleanMasks(groups []string) []string {
	out := make([]string, 0, len(groups))

	for _, m := range groups {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}

	return out
}

// loadSettingsRow picks up the platform settings that outlive a pod, so a restart does not
// quietly put one of them back to its default.
func (h *Handler) LoadSystemConfig(ctx context.Context) error {
	var at string

	err := h.DB.DB.QueryRow(ctx,
		`SELECT value FROM hst.settings WHERE key = 'end_of_day_at'`).Scan(&at)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	if at != "" {
		h.ChangeEndOfDayDate(at)
	}

	return nil
}

// LoadMarketData reads back the last price of every instrument, so the engine is not blind until
// each one next prints.
//
// It reads them out of the cache the feed keeps rather than asking the feed, because the moment
// this matters most is the moment the feed is down: a pod coming up on a weekend has no quotes
// coming and nobody to ask for them. The cache outlives both services.
//
// Nothing here is fatal. A pod that refuses to start is an outage; a pod with an empty book is
// only degraded, and every symbol heals on its first tick. A price older than the configured
// bound is left out rather than seeded: a missing symbol makes the book report nothing and the
// caller refuse, while a stale one would be taken for the market and could stop an account out
// on Monday against Friday's close.
func (h *Handler) LoadMarketData(ctx context.Context) error {
	symbols := h.Settings.SymbolNames()
	if len(symbols) == 0 || h.Redis == nil || h.Redis.Client == nil {
		return nil
	}

	keys := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		keys = append(keys, lastPriceKey+symbol)
	}

	values, err := h.Redis.Client.MGet(ctx, keys...).Result()
	if err != nil {
		h.Log.Log(logger.TypeCfg, logger.CodeWarn, "no last prices to start from",
			"error", err.Error())
		return nil
	}

	maxAge := h.priceMaxAge()

	var seeded, stale int

	for _, v := range values {
		text, ok := v.(string)
		if !ok {
			continue
		}

		var t model.Tick
		if err := json.Unmarshal([]byte(text), &t); err != nil || !t.Ok() {
			continue
		}

		if time.Since(time.Unix(0, t.Time)) > maxAge {
			stale++
			continue
		}

		h.Quotes.Set(t)
		seeded++
	}

	h.Log.Log(logger.TypeCfg, logger.CodeOK, "last prices loaded",
		"seeded", seeded, "too_old", stale, "max_age", maxAge.String())

	return nil
}

// lastPriceKey is where the feed leaves the last price of a symbol.
const lastPriceKey = "hstquote:last:"

// priceMaxAge is how old a price may be and still be worth starting from.
func (h *Handler) priceMaxAge() time.Duration {
	if h.Cfg.Engine.PriceMaxAge > 0 {
		return h.Cfg.Engine.PriceMaxAge
	}

	return 72 * time.Hour
}
