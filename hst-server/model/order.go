package model

// Order enums and record.

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

var OrderType_name = map[int32]string{
	0: "buy", 1: "sell", 2: "buy_limit", 3: "sell_limit", 4: "buy_stop",
	5: "sell_stop", 6: "buy_stop_limit", 7: "sell_stop_limit", 8: "close_by",
}

// IsPending reports whether the order waits for a price rather than filling now.
func (t OrderType) IsPending() bool {
	return t >= OrderType_buy_limit && t <= OrderType_sell_stop_limit
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

var OrderState_name = map[int32]string{
	0: "started", 1: "placed", 2: "canceled", 3: "partial", 4: "filled",
	5: "rejected", 6: "expired", 7: "request_add", 8: "request_modify", 9: "request_cancel",
}

// IsLive reports whether the order still sits on the book.
func (s OrderState) IsLive() bool {
	return s == OrderState_started || s == OrderState_placed || s == OrderState_partial
}

// OrderFilling is what to do when the book cannot fill the whole volume.
type OrderFilling int32

const (
	OrderFilling_fok    OrderFilling = 0 // all or none
	OrderFilling_ioc    OrderFilling = 1 // cancel the remainder
	OrderFilling_return OrderFilling = 2 // return the remainder to the book
	OrderFilling_boc    OrderFilling = 3 // book or cancel, passive only
)

var OrderFilling_name = map[int32]string{0: "fok", 1: "ioc", 2: "return", 3: "boc"}

// OrderTime is how long the order lives.
type OrderTime int32

const (
	OrderTime_gtc           OrderTime = 0
	OrderTime_day           OrderTime = 1
	OrderTime_specified     OrderTime = 2
	OrderTime_specified_day OrderTime = 3
)

var OrderTime_name = map[int32]string{0: "gtc", 1: "day", 2: "specified", 3: "specified_day"}

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

var OrderReason_name = map[int32]string{
	0: "client", 1: "expert", 2: "dealer", 3: "sl", 4: "tp", 5: "so", 6: "rollover",
	7: "external_client", 8: "vmargin", 9: "gateway", 10: "signal", 11: "settlement",
	12: "transfer", 13: "sync", 14: "external_service", 15: "migration", 16: "mobile",
	17: "web", 18: "split", 19: "corporate_action", 20: "ultency", 21: "coverage",
}

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

// TradeActivationFlags disables server-side activation, so an external system can own it instead.
type TradeActivationFlags int32

const (
	ActivationFlags_none      TradeActivationFlags = 0x00000000
	ActivationFlags_no_limit  TradeActivationFlags = 0x00000001
	ActivationFlags_no_stop   TradeActivationFlags = 0x00000002
	ActivationFlags_no_slimit TradeActivationFlags = 0x00000004
	ActivationFlags_no_sl     TradeActivationFlags = 0x00000008
	ActivationFlags_no_tp     TradeActivationFlags = 0x00000010
	ActivationFlags_no_so     TradeActivationFlags = 0x00000020
	ActivationFlags_no_expiry TradeActivationFlags = 0x00000040
)

// TradeModifyFlags records who edited the record by hand.
type TradeModifyFlags int32

const (
	ModifyFlags_none        TradeModifyFlags = 0x00000000
	ModifyFlags_admin       TradeModifyFlags = 0x00000001
	ModifyFlags_manager     TradeModifyFlags = 0x00000002
	ModifyFlags_restore     TradeModifyFlags = 0x00000004
	ModifyFlags_api_admin   TradeModifyFlags = 0x00000008
	ModifyFlags_api_manager TradeModifyFlags = 0x00000010
	ModifyFlags_api_server  TradeModifyFlags = 0x00000020
	ModifyFlags_api_gateway TradeModifyFlags = 0x00000040
)

// Order is one row of hst.orders. Volume is integer units, money is decimal, times are epoch nanoseconds.
type Order struct {
	OrderId          int64   `json:"order_id"`
	ExternalId       string  `json:"external_id"`
	Login            int64   `json:"login"`
	Dealer           int64   `json:"dealer"`
	Symbol           string  `json:"symbol"`
	Digits           int32   `json:"digits"`
	DigitsCurrency   int32   `json:"digits_currency"`
	ContractSize     float64 `json:"contract_size"`
	State            int32   `json:"state"`
	Reason           int32   `json:"reason"`
	TimeSetup        int64   `json:"time_setup"`
	TimeExpiration   int64   `json:"time_expiration"`
	TimeDone         int64   `json:"time_done"`
	ModifyFlags      int32   `json:"modify_flags"`
	Type             int32   `json:"type"`
	TypeFill         int32   `json:"type_fill"`
	TypeTime         int32   `json:"type_time"`
	PriceOrder       float64 `json:"price_order"`
	PriceTrigger     float64 `json:"price_trigger"`
	PriceCurrent     float64 `json:"price_current"`
	PriceSL          float64 `json:"price_sl"`
	PriceTP          float64 `json:"price_tp"`
	VolumeInitial    int64   `json:"volume_initial"`
	VolumeInitialExt int64   `json:"volume_initial_ext"`
	VolumeCurrent    int64   `json:"volume_current"`
	VolumeCurrentExt int64   `json:"volume_current_ext"`
	ExpertId         int64   `json:"expert_id"`
	PositionId       int64   `json:"position_id"`
	PositionById     int64   `json:"position_by_id"`
	Comment          string  `json:"comment"`
	ActivationMode   int32   `json:"activation_mode"`
	ActivationTime   int64   `json:"activation_time"`
	ActivationPrice  float64 `json:"activation_price"`
	ActivationFlags  int32   `json:"activation_flags"`
	RateMargin       float64 `json:"rate_margin"`
	ApiData          string  `json:"api_data"`
	DateCreated      int64   `json:"date_created"`
	DateModified     int64   `json:"date_modified"`
}

// VolumeLots is the order volume as a decimal number of lots.
func (o *Order) VolumeLots() float64 { return VolumeToLots(o.VolumeCurrent) }

// VolumeUnit is one volume unit: 1/10000 of a lot.
const VolumeUnit = 10000.0

// VolumeToLots converts integer volume into lots.
func VolumeToLots(v int64) float64 { return float64(v) / VolumeUnit }

// LotsToVolume converts lots into integer volume.
func LotsToVolume(lots float64) int64 { return int64(lots*VolumeUnit + 0.5) }
