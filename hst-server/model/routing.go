package model

// Routing enums.

var (
	RouteFlags_name = map[int32]string{
		0: "none", 0x00000001: "price", 0x00000002: "request", 0x00000004: "instant",
		0x00000008: "market", 0x00000010: "exchange", 0x00000020: "pending",
		0x00000040: "sltp", 0x00000080: "modify", 0x00000100: "remove",
		0x00000200: "activate", 0x00000400: "stoplimit", 0x00000800: "sl",
		0x00001000: "tp", 0x00002000: "stopout_order", 0x00004000: "stopout_position",
		0x00008000: "expiration", 0x00010000: "dealer_pos_execute",
		0x00020000: "dealer_ord_pending", 0x00040000: "dealer_pos_modify",
		0x00080000: "dealer_ord_modify", 0x00100000: "dealer_ord_remove",
		0x00200000: "dealer_ord_activate", 0x00400000: "dealer_ord_slimit",
		0x00800000: "dealer_close_by", 0x01000000: "close_by",
	}
)

var (
	TypeFlags_name = map[int32]string{
		0: "none", 0x0001: "buy", 0x0002: "sell", 0x0004: "buy_limit",
		0x0008: "sell_limit", 0x0010: "buy_stop", 0x0020: "sell_stop",
		0x0040: "buy_stop_limit", 0x0080: "sell_stop_limit",
	}
)

var (
	RouteAction_name = map[int32]string{
		0: "delay_time", 1: "delay_tick", 2: "clear_tp", 3: "clear_sl", 4: "clear_sltp",
		1001: "dealer", 1002: "dealer_online",
		1003: "reject", 1004: "requote", 1005: "confirm_client",
		1006: "confirm_market", 1007: "cancel_order",
	}
	RouteAction_value = map[string]int32{
		"delay_time": 0, "delay_tick": 1, "clear_tp": 2, "clear_sl": 3, "clear_sltp": 4,
		"dealer": 1001, "dealer_online": 1002,
		"reject": 1003, "requote": 1004, "confirm_client": 1005,
		"confirm_market": 1006, "cancel_order": 1007,
	}
)

var (
	RouteCondition_name = map[int32]string{
		0: "datetime", 1: "symbol", 2: "volume", 3: "market_deviation", 4: "time",
		5: "weekday", 6: "comment", 7: "expert", 8: "signal", 9: "dealer_login",
		10: "source_login", 11: "market_deviation_spr", 12: "gap",
		1000: "login", 1001: "group", 1002: "country", 1003: "city", 1004: "color",
		1005: "leverage", 1006: "comment_client",
		2000: "margin", 2001: "margin_level", 2002: "margin_free", 2003: "equity",
		2004: "balance", 2005: "profit",
		3000: "daily_deals", 3001: "daily_deals_period", 3002: "daily_profit",
		4000: "position_volume", 4001: "position_profit", 4002: "position_age",
		4003: "position_modify_time", 4004: "position_average_time",
		4005: "position_total", 4006: "position_total_symbol",
		4007: "order_total", 4008: "order_total_symbol",
		4009: "position_sl_touched", 4010: "position_tp_touched",
		4011: "order_sl_touched", 4012: "order_tp_touched",
		13: "reason", 14: "request_price", 15: "value", 16: "current_spread",
		1007: "zip", 1008: "status", 1009: "client_id", 1010: "party_id",
		4013: "position_value", 4014: "order_in", 4015: "order_out",
	}
)

var (
	ConditionRule_name = map[int32]string{
		0: "eq", 1: "not_eq", 2: "greater", 3: "not_less", 4: "less", 5: "not_greater",
	}
	ConditionRule_value = map[string]int32{
		"eq": 0, "not_eq": 1, "greater": 2, "not_less": 3, "less": 4, "not_greater": 5,
	}
)

type RoutingRule struct {
	RoutingId    int64       `db:"routing_id" json:"routing_id"`
	Name         string      `db:"name" json:"name"`
	Mode         int16       `db:"mode" json:"mode"`
	Request      int32       `db:"request" json:"request"`
	Type         int32       `db:"type" json:"type"`
	Flags        int32       `db:"flags" json:"flags"`
	Action       RouteAction `db:"action" json:"action"`
	ActionValue  string      `db:"action_value" json:"action_value"`
	RoutingIndex int32       `db:"routing_index" json:"routing_index"`
	DateCreated  int64       `db:"date_created" json:"date_created"`
	DateModified int64       `db:"date_modified" json:"date_modified"`
}

func (RoutingRule) TableName() string { return "hst.routing" }

type RoutingCondition struct {
	ConditionId int64          `db:"condition_id" json:"condition_id"`
	RoutingId   int64          `db:"routing_id" json:"routing_id"`
	Condition   RouteCondition `db:"condition" json:"condition"`
	Rule        ConditionRule  `db:"rule" json:"rule"`
	Value       string         `db:"value" json:"value"`
}

func (RoutingCondition) TableName() string { return "hst.routing_conds" }

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
