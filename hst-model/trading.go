package model

// How an instrument is priced, margined, executed and routed. Shared, because the engine and
// the api must agree on every one of these to agree on a trade.

// How the margin for an instrument is worked out.
type CalcMode int32

const (
	CalcMode_forex               CalcMode = 0
	CalcMode_futures             CalcMode = 1
	CalcMode_cfd                 CalcMode = 2
	CalcMode_cfd_index           CalcMode = 3
	CalcMode_cfd_leverage        CalcMode = 4
	CalcMode_forex_no_leverage   CalcMode = 5
	CalcMode_exch_stocks         CalcMode = 32
	CalcMode_exch_futures        CalcMode = 33
	CalcMode_exch_forts          CalcMode = 34
	CalcMode_exch_options        CalcMode = 35
	CalcMode_exch_options_margin CalcMode = 36
	CalcMode_exch_bonds          CalcMode = 37
	CalcMode_serv_collateral     CalcMode = 64
)

// What a group may do with an instrument.
type TradeMode int32

const (
	TradeMode_disabled   TradeMode = 0
	TradeMode_long_only  TradeMode = 1
	TradeMode_short_only TradeMode = 2
	TradeMode_close_only TradeMode = 3
	TradeMode_full       TradeMode = 4
)

// How a request is turned into a fill.
type ExecMode int32

const (
	ExecMode_request  ExecMode = 0
	ExecMode_instant  ExecMode = 1
	ExecMode_market   ExecMode = 2
	ExecMode_exchange ExecMode = 3
)

// Netting keeps one position per symbol; hedging lets many sit side by side.
type MarginMode int32

const (
	MarginMode_retail_netting MarginMode = 0
	MarginMode_exchange       MarginMode = 1
	MarginMode_retail_hedging MarginMode = 2
)

// How the stop out level is read.
type StopOutMode int32

const (
	StopOutMode_percent StopOutMode = 0
	StopOutMode_money   StopOutMode = 1
)

// How much of the floating result counts toward free margin, from hst.groups.margin_free_mode.
type FreeMarginMode int32

const (
	FreeMarginMode_not_use_pl FreeMarginMode = 0
	FreeMarginMode_use_pl     FreeMarginMode = 1
	FreeMarginMode_profit     FreeMarginMode = 2
	FreeMarginMode_loss       FreeMarginMode = 3
)

// How a day's realised profit is treated, from hst.groups.margin_free_profit_mode.
type FreeMarginProfitMode int32

const (
	FreeMarginProfitMode_day_profit_and_loss FreeMarginProfitMode = 0
	FreeMarginProfitMode_day_profit_loss     FreeMarginProfitMode = 1
)

// Which order types a rule applies to, as a set of flags. Zero means all of them.
type TypeFlags int32

const (
	TypeFlags_none            TypeFlags = 0
	TypeFlags_buy             TypeFlags = 0x0001
	TypeFlags_sell            TypeFlags = 0x0002
	TypeFlags_buy_limit       TypeFlags = 0x0004
	TypeFlags_sell_limit      TypeFlags = 0x0008
	TypeFlags_buy_stop        TypeFlags = 0x0010
	TypeFlags_sell_stop       TypeFlags = 0x0020
	TypeFlags_buy_stop_limit  TypeFlags = 0x0040
	TypeFlags_sell_stop_limit TypeFlags = 0x0080
)

// Which kinds of request a rule applies to, as a set of flags. Zero means all of them.
type RouteFlags int32

const (
	RouteFlags_none                RouteFlags = 0
	RouteFlags_price               RouteFlags = 0x00000001
	RouteFlags_request             RouteFlags = 0x00000002
	RouteFlags_instant             RouteFlags = 0x00000004
	RouteFlags_market              RouteFlags = 0x00000008
	RouteFlags_exchange            RouteFlags = 0x00000010
	RouteFlags_pending             RouteFlags = 0x00000020
	RouteFlags_sltp                RouteFlags = 0x00000040
	RouteFlags_modify              RouteFlags = 0x00000080
	RouteFlags_remove              RouteFlags = 0x00000100
	RouteFlags_activate            RouteFlags = 0x00000200
	RouteFlags_stop_limit          RouteFlags = 0x00000400
	RouteFlags_sl                  RouteFlags = 0x00000800
	RouteFlags_tp                  RouteFlags = 0x00001000
	RouteFlags_stop_out_order      RouteFlags = 0x00002000
	RouteFlags_stop_out_position   RouteFlags = 0x00004000
	RouteFlags_expiration          RouteFlags = 0x00008000
	RouteFlags_close_by            RouteFlags = 0x01000000
	RouteFlags_dealer_pos_execute  RouteFlags = 0x00010000
	RouteFlags_dealer_ord_pending  RouteFlags = 0x00020000
	RouteFlags_dealer_pos_modify   RouteFlags = 0x00040000
	RouteFlags_dealer_ord_modify   RouteFlags = 0x00080000
	RouteFlags_dealer_ord_remove   RouteFlags = 0x00100000
	RouteFlags_dealer_ord_activate RouteFlags = 0x00200000
	RouteFlags_dealer_ord_slimit   RouteFlags = 0x00400000
	RouteFlags_dealer_close_by     RouteFlags = 0x00800000
)

// What a rule does to a request that matches it.
type RouteAction int32

const (
	RouteAction_delay_time     RouteAction = 0
	RouteAction_delay_tick     RouteAction = 1
	RouteAction_clear_tp       RouteAction = 2
	RouteAction_clear_sl       RouteAction = 3
	RouteAction_clear_sltp     RouteAction = 4
	RouteAction_dealer         RouteAction = 1001
	RouteAction_dealer_online  RouteAction = 1002
	RouteAction_reject         RouteAction = 1003
	RouteAction_requote        RouteAction = 1004
	RouteAction_confirm_client RouteAction = 1005
	RouteAction_confirm_market RouteAction = 1006
	RouteAction_cancel_order   RouteAction = 1007
)

// What a rule's extra condition looks at.
type RouteCondition int32

const (
	RouteCondition_datetime         RouteCondition = 0
	RouteCondition_symbol           RouteCondition = 1
	RouteCondition_volume           RouteCondition = 2
	RouteCondition_deviation        RouteCondition = 3
	RouteCondition_time             RouteCondition = 4
	RouteCondition_weekday          RouteCondition = 5
	RouteCondition_comment          RouteCondition = 6
	RouteCondition_expert           RouteCondition = 7
	RouteCondition_signal           RouteCondition = 8
	RouteCondition_dealer_login     RouteCondition = 9
	RouteCondition_source_login     RouteCondition = 10
	RouteCondition_deviation_spread RouteCondition = 11
	RouteCondition_gap              RouteCondition = 12
	RouteCondition_reason           RouteCondition = 13
	RouteCondition_request_price    RouteCondition = 14
	RouteCondition_value            RouteCondition = 15
	RouteCondition_current_spread   RouteCondition = 16

	RouteCondition_login     RouteCondition = 1000
	RouteCondition_group     RouteCondition = 1001
	RouteCondition_country   RouteCondition = 1002
	RouteCondition_city      RouteCondition = 1003
	RouteCondition_color     RouteCondition = 1004
	RouteCondition_leverage  RouteCondition = 1005
	RouteCondition_comment2  RouteCondition = 1006
	RouteCondition_zip       RouteCondition = 1007
	RouteCondition_status    RouteCondition = 1008
	RouteCondition_client_id RouteCondition = 1009
	RouteCondition_party_id  RouteCondition = 1010

	RouteCondition_margin       RouteCondition = 2000
	RouteCondition_margin_level RouteCondition = 2001
	RouteCondition_margin_free  RouteCondition = 2002
	RouteCondition_equity       RouteCondition = 2003
	RouteCondition_balance      RouteCondition = 2004
	RouteCondition_profit       RouteCondition = 2005

	RouteCondition_daily_deals        RouteCondition = 3000
	RouteCondition_daily_deals_period RouteCondition = 3001
	RouteCondition_daily_profit       RouteCondition = 3002

	RouteCondition_position_volume       RouteCondition = 4000
	RouteCondition_position_profit       RouteCondition = 4001
	RouteCondition_position_age          RouteCondition = 4002
	RouteCondition_position_modify_time  RouteCondition = 4003
	RouteCondition_position_average_time RouteCondition = 4004
	RouteCondition_position_total        RouteCondition = 4005
	RouteCondition_position_total_symbol RouteCondition = 4006
	RouteCondition_order_total           RouteCondition = 4007
	RouteCondition_order_total_symbol    RouteCondition = 4008
	RouteCondition_position_sl_touched   RouteCondition = 4009
	RouteCondition_position_tp_touched   RouteCondition = 4010
	RouteCondition_order_sl_touched      RouteCondition = 4011
	RouteCondition_order_tp_touched      RouteCondition = 4012
	RouteCondition_position_value        RouteCondition = 4013
	RouteCondition_order_in              RouteCondition = 4014
	RouteCondition_order_out             RouteCondition = 4015
)

type QueryWhat int32

const (
	QueryWhat_account   QueryWhat = 1
	QueryWhat_positions QueryWhat = 2
	QueryWhat_orders    QueryWhat = 3
	QueryWhat_state     QueryWhat = 4
	QueryWhat_symbols   QueryWhat = 5
)

func (m TradeMode) AllowsBuy() bool { return m == TradeMode_full || m == TradeMode_long_only }

func (m TradeMode) AllowsSell() bool { return m == TradeMode_full || m == TradeMode_short_only }

func (m TradeMode) CloseOnly() bool { return m == TradeMode_close_only }

func (m MarginMode) Hedging() bool { return m == MarginMode_retail_hedging }

// A delay or a cleared level lets the request carry on; everything else settles it.
func (a RouteAction) Terminal() bool {
	switch a {
	case RouteAction_delay_time, RouteAction_delay_tick, RouteAction_clear_tp, RouteAction_clear_sl, RouteAction_clear_sltp:
		return false
	}
	return true
}

// ToDealer reports whether the action hands the request to the dealing desk.
func (a RouteAction) ToDealer() bool {
	return a == RouteAction_dealer || a == RouteAction_dealer_online
}

func (a RouteAction) Executes() bool {
	return a == RouteAction_confirm_client || a == RouteAction_confirm_market
}

// How a condition compares what it looked at against the rule's value.
type ConditionRule int16

const (
	ConditionRule_equal       ConditionRule = 0
	ConditionRule_not_equal   ConditionRule = 1
	ConditionRule_greater     ConditionRule = 2
	ConditionRule_not_less    ConditionRule = 3
	ConditionRule_less        ConditionRule = 4
	ConditionRule_not_greater ConditionRule = 5
)

// Counts reports how much of a floating result the mode admits into free margin.
func (m FreeMarginMode) Counts(floating float64) float64 {
	switch m {
	case FreeMarginMode_use_pl:
		return floating
	case FreeMarginMode_profit:
		if floating > 0 {
			return floating
		}
	case FreeMarginMode_loss:
		if floating < 0 {
			return floating
		}
	}

	return 0
}
