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
