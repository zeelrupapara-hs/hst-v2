package model

// One volume unit is 1/10000 of a lot, which is what the legacy volume columns hold.
const VolumeUnit = 10000.0

// The engine counts in extended units of 1/100000000 of a lot, so an instrument can be traded
// finer than a legacy unit allows. The legacy number is derived from this one, never the reverse.
const VolumeUnitExt = 100000000.0

// ExtPerUnit is how many extended units one legacy unit is worth.
const ExtPerUnit = int64(VolumeUnitExt / VolumeUnit)

func Lots(v int64) float64 { return float64(v) / VolumeUnitExt }

// Legacy is an extended volume as the coarser number the old columns carry.
func Legacy(ext int64) int64 { return ext / ExtPerUnit }

// FromLegacy reads a coarse volume as extended units, for a row written before the change.
func FromLegacy(v int64) int64 { return v * ExtPerUnit }

// Extended picks whichever of the two a row actually carries.
func Extended(volume, ext int64) int64 {
	if ext > 0 {
		return ext
	}

	return FromLegacy(volume)
}

type Order struct {
	OrderId        int64        `json:"order_id"`
	Login          int64        `json:"login"`
	Dealer         int64        `json:"dealer"`
	Symbol         string       `json:"symbol"`
	Digits         int32        `json:"digits"`
	DigitsCurrency int32        `json:"digits_currency"`
	ContractSize   float64      `json:"contract_size"`
	State          OrderState   `json:"state"`
	Reason         OrderReason  `json:"reason"`
	TimeSetup      int64        `json:"time_setup"`
	TimeExpiration int64        `json:"time_expiration"`
	TimeDone       int64        `json:"time_done"`
	Type           OrderType    `json:"type"`
	TypeFill       OrderFilling `json:"type_fill"`
	TypeTime       OrderTime    `json:"type_time"`
	PriceOrder     float64      `json:"price_order"`
	PriceTrigger   float64      `json:"price_trigger"`
	PriceCurrent   float64      `json:"price_current"`
	PriceSL        float64      `json:"price_sl"`
	PriceTP        float64      `json:"price_tp"`
	VolumeInitial  int64        `json:"volume_initial"`
	VolumeCurrent  int64        `json:"volume_current"`
	VolumeExt      int64        `json:"volume_ext"`
	ExpertId       int64        `json:"expert_id"`
	PositionId     int64        `json:"position_id"`
	PositionById   int64        `json:"position_by_id"`
	Comment        string       `json:"comment"`
	RateMargin     float64      `json:"rate_margin"`
	// RoutingId is the rule that sent this request to the dealing desk, if one did.
	RoutingId int64 `json:"routing_id"`

	ActivationMode  int32   `json:"activation_mode"`
	ActivationTime  int64   `json:"activation_time"`
	ActivationPrice float64 `json:"activation_price"`
	ActivationFlags int32   `json:"activation_flags"`
}

// Why an order was last touched by the price.
const (
	ActivationNone      = 0
	ActivationPending   = 1
	ActivationStopLimit = 2
	ActivationSL        = 3
	ActivationTP        = 4
	ActivationStopOut   = 5
)

// Which levels the engine must leave alone because something upstream owns them.
const (
	ActivationFlagNone        int32 = 0x00
	ActivationFlagNoExpiry    int32 = 0x01
	ActivationFlagNoSL        int32 = 0x02
	ActivationFlagNoTP        int32 = 0x04
	ActivationFlagNoStopLimit int32 = 0x08
)

func (o *Order) Lots() float64 { return Lots(o.VolumeCurrent) }

func (o *Order) Kind() OrderType { return o.Type }

type Position struct {
	PositionId     int64          `json:"position_id"`
	Login          int64          `json:"login"`
	Dealer         int64          `json:"dealer"`
	Symbol         string         `json:"symbol"`
	Action         PositionAction `json:"action"`
	Digits         int32          `json:"digits"`
	DigitsCurrency int32          `json:"digits_currency"`
	Reason         OrderReason    `json:"reason"`
	ContractSize   float64        `json:"contract_size"`
	TimeCreate     int64          `json:"time_create"`
	TimeUpdate     int64          `json:"time_update"`
	PriceOpen      float64        `json:"price_open"`
	PriceCurrent   float64        `json:"price_current"`
	PriceSL        float64        `json:"price_sl"`
	PriceTP        float64        `json:"price_tp"`
	Volume         int64          `json:"volume"`
	VolumeExt      int64          `json:"volume_ext"`
	Profit         float64        `json:"profit"`
	Storage        float64        `json:"storage"`
	RateProfit     float64        `json:"rate_profit"`
	RateMargin     float64        `json:"rate_margin"`
	ExpertId       int64          `json:"expert_id"`
	Comment        string         `json:"comment"`

	ActivationFlags int32 `json:"activation_flags"`

	// Held in memory only: derived from the price and the settings.
	Margin float64 `json:"margin"`
}

func (p *Position) IsBuy() bool { return p.Action.IsBuy() }

func (p *Position) Lots() float64 { return Lots(p.Volume) }

type Deal struct {
	DealId         int64       `json:"deal_id"`
	Login          int64       `json:"login"`
	Dealer         int64       `json:"dealer"`
	OrderId        int64       `json:"order_id"`
	Action         DealAction  `json:"action"`
	Entry          DealEntry   `json:"entry"`
	Digits         int32       `json:"digits"`
	DigitsCurrency int32       `json:"digits_currency"`
	ContractSize   float64     `json:"contract_size"`
	Time           int64       `json:"time"`
	Symbol         string      `json:"symbol"`
	Price          float64     `json:"price"`
	PriceSL        float64     `json:"price_sl"`
	PriceTP        float64     `json:"price_tp"`
	Volume         int64       `json:"volume"`
	VolumeExt      int64       `json:"volume_ext"`
	VolumeClosed   int64       `json:"volume_closed"`
	Profit         float64     `json:"profit"`
	Value          float64     `json:"value"`
	Storage        float64     `json:"storage"`
	Commission     float64     `json:"commission"`
	Fee            float64     `json:"fee"`
	RateProfit     float64     `json:"rate_profit"`
	RateMargin     float64     `json:"rate_margin"`
	ExpertId       int64       `json:"expert_id"`
	PositionId     int64       `json:"position_id"`
	Comment        string      `json:"comment"`
	ProfitRaw      float64     `json:"profit_raw"`
	PricePosition  float64     `json:"price_position"`
	TickValue      float64     `json:"tick_value"`
	TickSize       float64     `json:"tick_size"`
	Reason         OrderReason `json:"reason"`
	MarketBid      float64     `json:"market_bid"`
	MarketAsk      float64     `json:"market_ask"`
}

type Account struct {
	Login          int64  `json:"login"`
	Group          string `json:"group"`
	Currency       string `json:"currency"`
	CurrencyDigits int32  `json:"currency_digits"`
	Leverage       int32  `json:"leverage"`

	Balance           float64 `json:"balance"`
	Credit            float64 `json:"credit"`
	Margin            float64 `json:"margin"`
	MarginFree        float64 `json:"margin_free"`
	MarginLevel       float64 `json:"margin_level"`
	MarginInitial     float64 `json:"margin_initial"`
	MarginMaintenance float64 `json:"margin_maintenance"`
	Profit            float64 `json:"profit"`
	Storage           float64 `json:"storage"`
	Floating          float64 `json:"floating"`
	Equity            float64 `json:"equity"`
	Assets            float64 `json:"assets"`
	Liabilities       float64 `json:"liabilities"`
	Commission        float64 `json:"blocked_commission"`
	BlockedProfit     float64 `json:"blocked_profit"`
	UpdatedAt         int64   `json:"updated_at"`
	// VirtualCredit comes from the group and backs margin without ever being withdrawable.
	VirtualCredit float64 `json:"virtual_credit"`

	Rights int64 `json:"rights"`
}

type Direction int32

const (
	Direction_in  Direction = 0
	Direction_out Direction = 1
)

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
		0: "buy",
		1: "sell",
		2: "buy_limit",
		3: "sell_limit",
		4: "buy_stop",
		5: "sell_stop",
		6: "buy_stop_limit",
		7: "sell_stop_limit",
		8: "close_by",
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
		0: "started",
		1: "placed",
		2: "canceled",
		3: "partial",
		4: "filled",
		5: "rejected",
		6: "expired",
		7: "request_add",
		8: "request_modify",
		9: "request_cancel",
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
	OrderFilling_name = map[int32]string{
		0: "fok",
		1: "ioc",
		2: "return",
		3: "boc",
	}
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
	OrderTime_name = map[int32]string{
		0: "gtc",
		1: "day",
		2: "specified",
		3: "specified_day",
	}
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
		0:  "client",
		1:  "expert",
		2:  "dealer",
		3:  "sl",
		4:  "tp",
		5:  "so",
		6:  "rollover",
		7:  "external_client",
		8:  "vmargin",
		9:  "gateway",
		10: "signal",
		11: "settlement",
		12: "transfer",
		13: "sync",
		14: "external_service",
		15: "migration",
		16: "mobile",
		17: "web",
		18: "split",
		19: "corporate_action",
		20: "ultency",
		21: "coverage",
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
		0: "none",
		1: "pending",
		2: "stoplimit",
		3: "sl",
		4: "tp",
		5: "stopout",
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

// PositionAction is the side a position holds. A position is only ever buy or sell.
type PositionAction int32

const (
	PositionAction_buy  PositionAction = 0
	PositionAction_sell PositionAction = 1
)

var (
	PositionAction_name = map[int32]string{
		0: "buy",
		1: "sell",
	}
	PositionAction_value = valuesOf(PositionAction_name)
)

// IsBuy reports which way the position leans.
func (a PositionAction) IsBuy() bool { return a == PositionAction_buy }

// DealAction is what the deal did to the account. Buy and sell trade; the rest move money.
type DealAction int32

const (
	DealAction_buy                    DealAction = 0
	DealAction_sell                   DealAction = 1
	DealAction_balance                DealAction = 2
	DealAction_credit                 DealAction = 3
	DealAction_charge                 DealAction = 4
	DealAction_correction             DealAction = 5
	DealAction_bonus                  DealAction = 6
	DealAction_commission             DealAction = 7
	DealAction_commission_daily       DealAction = 8
	DealAction_commission_monthly     DealAction = 9
	DealAction_agent_daily            DealAction = 10
	DealAction_agent_monthly          DealAction = 11
	DealAction_interest               DealAction = 12
	DealAction_buy_canceled           DealAction = 13
	DealAction_sell_canceled          DealAction = 14
	DealAction_dividend               DealAction = 15
	DealAction_dividend_franked       DealAction = 16
	DealAction_tax                    DealAction = 17
	DealAction_agent                  DealAction = 18
	DealAction_so_compensation        DealAction = 19
	DealAction_so_compensation_credit DealAction = 20
)

var (
	DealAction_name = map[int32]string{
		0:  "buy",
		1:  "sell",
		2:  "balance",
		3:  "credit",
		4:  "charge",
		5:  "correction",
		6:  "bonus",
		7:  "commission",
		8:  "commission_daily",
		9:  "commission_monthly",
		10: "agent_daily",
		11: "agent_monthly",
		12: "interest",
		13: "buy_canceled",
		14: "sell_canceled",
		15: "dividend",
		16: "dividend_franked",
		17: "tax",
		18: "agent",
		19: "so_compensation",
		20: "so_compensation_credit",
	}
	DealAction_value = valuesOf(DealAction_name)
)

// DealEntry says which way the deal moved the position.
type DealEntry int32

const (
	DealEntry_in     DealEntry = 0
	DealEntry_out    DealEntry = 1
	DealEntry_inout  DealEntry = 2
	DealEntry_out_by DealEntry = 3
)

var (
	DealEntry_name = map[int32]string{
		0: "in",
		1: "out",
		2: "inout",
		3: "out_by",
	}
	DealEntry_value = valuesOf(DealEntry_name)
)

// valuesOf inverts a _name map, so the two can never disagree.
func valuesOf(names map[int32]string) map[string]int32 {
	out := make(map[string]int32, len(names))
	for v, n := range names {
		out[n] = v
	}

	return out
}
