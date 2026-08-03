package model

// The rule list is a whitelist: a request that reaches the end without a terminal action is not processed.

// Which kinds of request a rule applies to, as a set of flags. Zero means all of them.
type RouteFlags int32

const (
	RouteNone            RouteFlags = 0
	RoutePrice           RouteFlags = 0x00000001
	RouteRequest         RouteFlags = 0x00000002
	RouteInstant         RouteFlags = 0x00000004
	RouteMarket          RouteFlags = 0x00000008
	RouteExchange        RouteFlags = 0x00000010
	RoutePending         RouteFlags = 0x00000020
	RouteSLTP            RouteFlags = 0x00000040
	RouteModify          RouteFlags = 0x00000080
	RouteRemove          RouteFlags = 0x00000100
	RouteActivate        RouteFlags = 0x00000200
	RouteStopLimit       RouteFlags = 0x00000400
	RouteSL              RouteFlags = 0x00000800
	RouteTP              RouteFlags = 0x00001000
	RouteStopOutOrder    RouteFlags = 0x00002000
	RouteStopOutPosition RouteFlags = 0x00004000
	RouteExpiration      RouteFlags = 0x00008000
	RouteCloseBy         RouteFlags = 0x01000000
)

// Which order types a rule applies to, as a set of flags. Zero means all of them.
type TypeFlags int32

const (
	TypeNone          TypeFlags = 0
	TypeBuy           TypeFlags = 0x0001
	TypeSell          TypeFlags = 0x0002
	TypeBuyLimit      TypeFlags = 0x0004
	TypeSellLimit     TypeFlags = 0x0008
	TypeBuyStop       TypeFlags = 0x0010
	TypeSellStop      TypeFlags = 0x0020
	TypeBuyStopLimit  TypeFlags = 0x0040
	TypeSellStopLimit TypeFlags = 0x0080
)

func TypeFlagFor(t OrderType) TypeFlags {
	switch t {
	case OrderType_buy:
		return TypeBuy
	case OrderType_sell:
		return TypeSell
	case OrderType_buy_limit:
		return TypeBuyLimit
	case OrderType_sell_limit:
		return TypeSellLimit
	case OrderType_buy_stop:
		return TypeBuyStop
	case OrderType_sell_stop:
		return TypeSellStop
	case OrderType_buy_stop_limit:
		return TypeBuyStopLimit
	case OrderType_sell_stop_limit:
		return TypeSellStopLimit
	}
	return TypeNone
}

// What a rule does to a request that matches it.
type RouteAction int32

const (
	ActionDelayTime     RouteAction = 0
	ActionDelayTick     RouteAction = 1
	ActionClearTP       RouteAction = 2
	ActionClearSL       RouteAction = 3
	ActionClearSLTP     RouteAction = 4
	ActionDealer        RouteAction = 1001
	ActionDealerOnline  RouteAction = 1002
	ActionReject        RouteAction = 1003
	ActionRequote       RouteAction = 1004
	ActionConfirmClient RouteAction = 1005
	ActionConfirmMarket RouteAction = 1006
	ActionCancelOrder   RouteAction = 1007
)

// A delay or a cleared level lets the request carry on; everything else settles it.
func (a RouteAction) Terminal() bool {
	switch a {
	case ActionDelayTime, ActionDelayTick, ActionClearTP, ActionClearSL, ActionClearSLTP:
		return false
	}
	return true
}

// ToDealer reports whether the action hands the request to the dealing desk.
func (a RouteAction) ToDealer() bool {
	return a == ActionDealer || a == ActionDealerOnline
}

func (a RouteAction) Executes() bool {
	return a == ActionConfirmClient || a == ActionConfirmMarket
}

// What a rule's extra condition looks at.
type RouteCondition int32

const (
	CondDatetime        RouteCondition = 0
	CondSymbol          RouteCondition = 1
	CondVolume          RouteCondition = 2
	CondDeviation       RouteCondition = 3
	CondTime            RouteCondition = 4
	CondWeekday         RouteCondition = 5
	CondComment         RouteCondition = 6
	CondExpert          RouteCondition = 7
	CondSignal          RouteCondition = 8
	CondDealerLogin     RouteCondition = 9
	CondSourceLogin     RouteCondition = 10
	CondDeviationSpread RouteCondition = 11
	CondGap             RouteCondition = 12
	CondReason          RouteCondition = 13
	CondRequestPrice    RouteCondition = 14
	CondValue           RouteCondition = 15
	CondCurrentSpread   RouteCondition = 16

	CondLogin    RouteCondition = 1000
	CondGroup    RouteCondition = 1001
	CondCountry  RouteCondition = 1002
	CondCity     RouteCondition = 1003
	CondColor    RouteCondition = 1004
	CondLeverage RouteCondition = 1005
	CondComment2 RouteCondition = 1006
	CondZip      RouteCondition = 1007
	CondStatus   RouteCondition = 1008
	CondClientId RouteCondition = 1009
	CondPartyId  RouteCondition = 1010

	CondMargin      RouteCondition = 2000
	CondMarginLevel RouteCondition = 2001
	CondMarginFree  RouteCondition = 2002
	CondEquity      RouteCondition = 2003
	CondBalance     RouteCondition = 2004
	CondProfit      RouteCondition = 2005

	CondDailyDeals       RouteCondition = 3000
	CondDailyDealsPeriod RouteCondition = 3001
	CondDailyProfit      RouteCondition = 3002

	CondPositionVolume      RouteCondition = 4000
	CondPositionProfit      RouteCondition = 4001
	CondPositionAge         RouteCondition = 4002
	CondPositionModifyTime  RouteCondition = 4003
	CondPositionAverageTime RouteCondition = 4004
	CondPositionTotal       RouteCondition = 4005
	CondPositionTotalSymbol RouteCondition = 4006
	CondOrderTotal          RouteCondition = 4007
	CondOrderTotalSymbol    RouteCondition = 4008
	CondPositionSLTouched   RouteCondition = 4009
	CondPositionTPTouched   RouteCondition = 4010
	CondOrderSLTouched      RouteCondition = 4011
	CondOrderTPTouched      RouteCondition = 4012
	CondPositionValue       RouteCondition = 4013
	CondOrderIn             RouteCondition = 4014
	CondOrderOut            RouteCondition = 4015
)

// How a condition compares what it looked at against the rule's value.
type ConditionRule int16

const (
	RuleEqual      ConditionRule = 0
	RuleNotEqual   ConditionRule = 1
	RuleGreater    ConditionRule = 2
	RuleNotLess    ConditionRule = 3
	RuleLess       ConditionRule = 4
	RuleNotGreater ConditionRule = 5
)

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
