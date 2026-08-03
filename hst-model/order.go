package model

// OrderType is what the client asked for.
type OrderType int32

const (
	OrderType_buy             OrderType = 0
	OrderType_sell            OrderType = 1
	OrderType_buy_limit       OrderType = 2
	OrderType_sell_limit      OrderType = 3
	OrderType_buy_stop        OrderType = 4
	OrderType_sell_stop       OrderType = 5
	OrderType_buy_stop_limit  OrderType = 6
	OrderType_sell_stop_limit OrderType = 7
	OrderType_close_by        OrderType = 8
)

var (
	OrderType_name = map[int32]string{
		0: "buy", 1: "sell", 2: "buy_limit", 3: "sell_limit", 4: "buy_stop",
		5: "sell_stop", 6: "buy_stop_limit", 7: "sell_stop_limit", 8: "close_by",
	}
	OrderType_value = valuesOf(OrderType_name)
)

// IsPending reports whether the order waits for a price rather than filling now.
func (t OrderType) IsPending() bool {
	return t >= OrderType_buy_limit && t <= OrderType_sell_stop_limit
}

// IsBuy reports which side of the book the order takes.
func (t OrderType) IsBuy() bool {
	switch t {
	case OrderType_buy, OrderType_buy_limit, OrderType_buy_stop, OrderType_buy_stop_limit:
		return true
	}

	return false
}

// IsMarket reports whether the order fills at once.
func (t OrderType) IsMarket() bool {
	return t == OrderType_buy || t == OrderType_sell
}

// OrderState is where the order is in its life.
type OrderState int32

const (
	OrderState_started        OrderState = 0
	OrderState_placed         OrderState = 1
	OrderState_canceled       OrderState = 2
	OrderState_partial        OrderState = 3
	OrderState_filled         OrderState = 4
	OrderState_rejected       OrderState = 5
	OrderState_expired        OrderState = 6
	OrderState_request_add    OrderState = 7
	OrderState_request_modify OrderState = 8
	OrderState_request_cancel OrderState = 9
)

var (
	OrderState_name = map[int32]string{
		0: "started", 1: "placed", 2: "canceled", 3: "partial", 4: "filled",
		5: "rejected", 6: "expired", 7: "request_add", 8: "request_modify", 9: "request_cancel",
	}
	OrderState_value = valuesOf(OrderState_name)
)

// IsLive reports whether the order still sits on the book. A queued request still cooks and
// still expires, so it counts.
func (s OrderState) IsLive() bool {
	switch s {
	case OrderState_started, OrderState_placed, OrderState_partial,
		OrderState_request_add, OrderState_request_modify, OrderState_request_cancel:
		return true
	}

	return false
}

// IsAwaitingDealer reports whether the order is sitting on a dealing desk.
func (s OrderState) IsAwaitingDealer() bool {
	switch s {
	case OrderState_request_add, OrderState_request_modify, OrderState_request_cancel:
		return true
	}

	return false
}

// OrderFilling is what to do when the book cannot fill the whole volume.
type OrderFilling int32

const (
	OrderFilling_fok    OrderFilling = 0
	OrderFilling_ioc    OrderFilling = 1
	OrderFilling_return OrderFilling = 2
	OrderFilling_boc    OrderFilling = 3
)

var (
	OrderFilling_name  = map[int32]string{0: "fok", 1: "ioc", 2: "return", 3: "boc"}
	OrderFilling_value = valuesOf(OrderFilling_name)
)

// OrderTime is how long the order lives. The zero value is good till cancelled.
type OrderTime int32

const (
	OrderTime_gtc           OrderTime = 0
	OrderTime_day           OrderTime = 1
	OrderTime_specified     OrderTime = 2
	OrderTime_specified_day OrderTime = 3
)

var (
	OrderTime_name  = map[int32]string{0: "gtc", 1: "day", 2: "specified", 3: "specified_day"}
	OrderTime_value = valuesOf(OrderTime_name)
)

// NeedsExpiry reports whether this lifetime requires the caller to name a moment.
func (t OrderTime) NeedsExpiry() bool {
	return t == OrderTime_specified || t == OrderTime_specified_day
}

// OrderReason is who or what placed the order.
type OrderReason int32

const (
	OrderReason_client           OrderReason = 0
	OrderReason_expert           OrderReason = 1
	OrderReason_dealer           OrderReason = 2
	OrderReason_sl               OrderReason = 3
	OrderReason_tp               OrderReason = 4
	OrderReason_so               OrderReason = 5
	OrderReason_rollover         OrderReason = 6
	OrderReason_external_client  OrderReason = 7
	OrderReason_vmargin          OrderReason = 8
	OrderReason_gateway          OrderReason = 9
	OrderReason_signal           OrderReason = 10
	OrderReason_settlement       OrderReason = 11
	OrderReason_transfer         OrderReason = 12
	OrderReason_sync             OrderReason = 13
	OrderReason_external_service OrderReason = 14
	OrderReason_migration        OrderReason = 15
	OrderReason_mobile           OrderReason = 16
	OrderReason_web              OrderReason = 17
	OrderReason_split            OrderReason = 18
	OrderReason_corporate_action OrderReason = 19
	OrderReason_ultency          OrderReason = 20
	OrderReason_coverage         OrderReason = 21
)

var (
	OrderReason_name = map[int32]string{
		0: "client", 1: "expert", 2: "dealer", 3: "sl", 4: "tp", 5: "so", 6: "rollover",
		7: "external_client", 8: "vmargin", 9: "gateway", 10: "signal", 11: "settlement",
		12: "transfer", 13: "sync", 14: "external_service", 15: "migration", 16: "mobile",
		17: "web", 18: "split", 19: "corporate_action", 20: "ultency", 21: "coverage",
	}
	OrderReason_value = valuesOf(OrderReason_name)
)

// OrderActivation is why the order was last touched by the price.
type OrderActivation int32

const (
	OrderActivation_none      OrderActivation = 0
	OrderActivation_pending   OrderActivation = 1
	OrderActivation_stoplimit OrderActivation = 2
	OrderActivation_sl        OrderActivation = 3
	OrderActivation_tp        OrderActivation = 4
	OrderActivation_stopout   OrderActivation = 5
)

var (
	OrderActivation_name = map[int32]string{
		0: "none", 1: "pending", 2: "stoplimit", 3: "sl", 4: "tp", 5: "stopout",
	}
	OrderActivation_value = valuesOf(OrderActivation_name)
)

// ActivationFlags disables server side activation, so an external system can own it instead.
type ActivationFlags int32

const (
	ActivationFlags_none      ActivationFlags = 0x00000000
	ActivationFlags_no_limit  ActivationFlags = 0x00000001
	ActivationFlags_no_stop   ActivationFlags = 0x00000002
	ActivationFlags_no_slimit ActivationFlags = 0x00000004
	ActivationFlags_no_sl     ActivationFlags = 0x00000008
	ActivationFlags_no_tp     ActivationFlags = 0x00000010
	ActivationFlags_no_so     ActivationFlags = 0x00000020
	ActivationFlags_no_expiry ActivationFlags = 0x00000040
)

// ModifyFlags records who edited the record by hand.
type ModifyFlags int32

const (
	ModifyFlags_none        ModifyFlags = 0x00000000
	ModifyFlags_admin       ModifyFlags = 0x00000001
	ModifyFlags_manager     ModifyFlags = 0x00000002
	ModifyFlags_restore     ModifyFlags = 0x00000004
	ModifyFlags_api_admin   ModifyFlags = 0x00000008
	ModifyFlags_api_manager ModifyFlags = 0x00000010
	ModifyFlags_api_server  ModifyFlags = 0x00000020
	ModifyFlags_api_gateway ModifyFlags = 0x00000040
)

// valuesOf inverts a _name map, so the two can never disagree.
func valuesOf(names map[int32]string) map[string]int32 {
	out := make(map[string]int32, len(names))
	for v, n := range names {
		out[n] = v
	}

	return out
}

// Valid reports whether a value is a member of its enum.
func Valid[T ~int32](v T, names map[int32]string) bool {
	_, ok := names[int32(v)]

	return ok
}
