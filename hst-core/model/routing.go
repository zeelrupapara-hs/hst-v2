package model

// The rule list is a whitelist: a request that reaches the end without a terminal action is not processed.

func TypeFlagFor(t OrderType) TypeFlags {
	switch t {
	case OrderType_buy:
		return TypeFlags_buy
	case OrderType_sell:
		return TypeFlags_sell
	case OrderType_buy_limit:
		return TypeFlags_buy_limit
	case OrderType_sell_limit:
		return TypeFlags_sell_limit
	case OrderType_buy_stop:
		return TypeFlags_buy_stop
	case OrderType_sell_stop:
		return TypeFlags_sell_stop
	case OrderType_buy_stop_limit:
		return TypeFlags_buy_stop_limit
	case OrderType_sell_stop_limit:
		return TypeFlags_sell_stop_limit
	}
	return TypeFlags_none
}

// RoutingRule is one line of the rule list.
type RoutingRule struct {
	RoutingId   int64
	Name        string
	Mode        int16
	Request     int32
	Type        int32
	Flags       int32
	Action      int32
	ActionValue string
	Index       int32

	Conditions []RoutingCondition
	Dealers    []int64
}

func (r *RoutingRule) Enabled() bool { return r.Mode == 1 }

// Every condition on a rule must hold for the rule to match.
type RoutingCondition struct {
	ConditionId int64
	Condition   int32
	Rule        int16
	Value       string
}
