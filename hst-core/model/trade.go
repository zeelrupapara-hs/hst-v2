package model

// Orders, positions and deals. Field for field from the MT5 SQL export, the same shape the
// hst.orders, hst.positions and hst.deals tables hold.

// VolumeUnit is one MT5 volume unit: 1/10000 of a lot. Volume is an integer everywhere, so a
// lot size can be compared and summed without the rounding a float would bring.
const VolumeUnit = 10000.0

// Lots turns MT5 integer volume into lots.
func Lots(v int64) float64 { return float64(v) / VolumeUnit }

// Volume turns lots into MT5 integer volume.
func Volume(lots float64) int64 { return int64(lots*VolumeUnit + 0.5) }

// What the client asked for.
type OrderType int32

const (
	OrderBuy           OrderType = 0
	OrderSell          OrderType = 1
	OrderBuyLimit      OrderType = 2
	OrderSellLimit     OrderType = 3
	OrderBuyStop       OrderType = 4
	OrderSellStop      OrderType = 5
	OrderBuyStopLimit  OrderType = 6
	OrderSellStopLimit OrderType = 7
	OrderCloseBy       OrderType = 8
)

// Buy reports whether the order takes the buy side.
func (t OrderType) Buy() bool {
	switch t {
	case OrderBuy, OrderBuyLimit, OrderBuyStop, OrderBuyStopLimit:
		return true
	}
	return false
}

// Pending reports whether the order waits for a price instead of filling now.
func (t OrderType) Pending() bool { return t >= OrderBuyLimit && t <= OrderSellStopLimit }

// Market reports whether the order fills immediately.
func (t OrderType) Market() bool { return t == OrderBuy || t == OrderSell }

// Where an order is in its life.
type OrderState int32

const (
	StateStarted  OrderState = 0
	StatePlaced   OrderState = 1
	StateCanceled OrderState = 2
	StatePartial  OrderState = 3
	StateFilled   OrderState = 4
	StateRejected OrderState = 5
	StateExpired  OrderState = 6
)

// Live reports whether the order is still working.
func (s OrderState) Live() bool {
	return s == StateStarted || s == StatePlaced || s == StatePartial
}

// What to do when the whole volume cannot be filled.
type Filling int32

const (
	FillFOK    Filling = 0
	FillIOC    Filling = 1
	FillReturn Filling = 2
	FillBOC    Filling = 3
)

// How long an order lives.
type Expiry int32

const (
	ExpiryGTC          Expiry = 0
	ExpiryDay          Expiry = 1
	ExpirySpecified    Expiry = 2
	ExpirySpecifiedDay Expiry = 3
)

// Who or what caused an order.
type Reason int32

const (
	ReasonClient   Reason = 0
	ReasonExpert   Reason = 1
	ReasonDealer   Reason = 2
	ReasonSL       Reason = 3
	ReasonTP       Reason = 4
	ReasonStopOut  Reason = 5
	ReasonRollover Reason = 6
	ReasonMobile   Reason = 16
	ReasonWeb      Reason = 17
)

// Order is one row of hst.orders.
type Order struct {
	OrderId        int64   `json:"order_id"`
	Login          int64   `json:"login"`
	Dealer         int64   `json:"dealer"`
	Symbol         string  `json:"symbol"`
	Digits         int32   `json:"digits"`
	DigitsCurrency int32   `json:"digits_currency"`
	ContractSize   float64 `json:"contract_size"`
	State          int32   `json:"state"`
	Reason         int32   `json:"reason"`
	TimeSetup      int64   `json:"time_setup"`
	TimeExpiration int64   `json:"time_expiration"`
	TimeDone       int64   `json:"time_done"`
	Type           int32   `json:"type"`
	TypeFill       int32   `json:"type_fill"`
	TypeTime       int32   `json:"type_time"`
	PriceOrder     float64 `json:"price_order"`
	PriceTrigger   float64 `json:"price_trigger"`
	PriceCurrent   float64 `json:"price_current"`
	PriceSL        float64 `json:"price_sl"`
	PriceTP        float64 `json:"price_tp"`
	VolumeInitial  int64   `json:"volume_initial"`
	VolumeCurrent  int64   `json:"volume_current"`
	ExpertId       int64   `json:"expert_id"`
	PositionId     int64   `json:"position_id"`
	PositionById   int64   `json:"position_by_id"`
	Comment        string  `json:"comment"`
	RateMargin     float64 `json:"rate_margin"`

	// Set when a stop limit's trigger is reached and it becomes a limit order.
	ActivationMode  int32   `json:"activation_mode"`
	ActivationTime  int64   `json:"activation_time"`
	ActivationPrice float64 `json:"activation_price"`
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

// Lots is the working volume as a decimal number of lots.
func (o *Order) Lots() float64 { return Lots(o.VolumeCurrent) }

// Kind is the order's type.
func (o *Order) Kind() OrderType { return OrderType(o.Type) }

// Position is one row of hst.positions.
//
// PriceOpen is the weighted average: a netting position grown by a second deal carries the
// blend of both, not the newer price.
type Position struct {
	PositionId     int64   `json:"position_id"`
	Login          int64   `json:"login"`
	Dealer         int64   `json:"dealer"`
	Symbol         string  `json:"symbol"`
	Action         int32   `json:"action"`
	Digits         int32   `json:"digits"`
	DigitsCurrency int32   `json:"digits_currency"`
	Reason         int32   `json:"reason"`
	ContractSize   float64 `json:"contract_size"`
	TimeCreate     int64   `json:"time_create"`
	TimeUpdate     int64   `json:"time_update"`
	PriceOpen      float64 `json:"price_open"`
	PriceCurrent   float64 `json:"price_current"`
	PriceSL        float64 `json:"price_sl"`
	PriceTP        float64 `json:"price_tp"`
	Volume         int64   `json:"volume"`
	Profit         float64 `json:"profit"`
	Storage        float64 `json:"storage"`
	RateProfit     float64 `json:"rate_profit"`
	RateMargin     float64 `json:"rate_margin"`
	ExpertId       int64   `json:"expert_id"`
	Comment        string  `json:"comment"`

	// Margin is what this position currently reserves. Held in memory only: it is derived from
	// the price and the settings, and recomputing it is cheaper than keeping it in step on disk.
	Margin float64 `json:"margin"`
}

// Buy reports whether the position is long.
func (p *Position) Buy() bool { return p.Action == 0 }

// Lots is the position volume as a decimal number of lots.
func (p *Position) Lots() float64 { return Lots(p.Volume) }

// What a deal did. Most values are not trades but balance movements.
type DealAction int32

const (
	DealBuy             DealAction = 0
	DealSell            DealAction = 1
	DealBalance         DealAction = 2
	DealCredit          DealAction = 3
	DealCharge          DealAction = 4
	DealCorrection      DealAction = 5
	DealBonus           DealAction = 6
	DealCommission      DealAction = 7
	DealCommissionDaily DealAction = 8
	DealInterest        DealAction = 12
	DealSOCompensation  DealAction = 19
)

// Which way a deal moved the position.
type DealEntry int32

const (
	EntryIn    DealEntry = 0 // opened or grew
	EntryOut   DealEntry = 1 // closed or shrank
	EntryInOut DealEntry = 2 // reversed
	EntryOutBy DealEntry = 3 // closed against an opposite position
)

// Deal is one row of hst.deals. This is the ledger: every money movement is a row here.
type Deal struct {
	DealId         int64   `json:"deal_id"`
	Login          int64   `json:"login"`
	Dealer         int64   `json:"dealer"`
	OrderId        int64   `json:"order_id"`
	Action         int32   `json:"action"`
	Entry          int32   `json:"entry"`
	Digits         int32   `json:"digits"`
	DigitsCurrency int32   `json:"digits_currency"`
	ContractSize   float64 `json:"contract_size"`
	Time           int64   `json:"time"`
	Symbol         string  `json:"symbol"`
	Price          float64 `json:"price"`
	PriceSL        float64 `json:"price_sl"`
	PriceTP        float64 `json:"price_tp"`
	Volume         int64   `json:"volume"`
	VolumeClosed   int64   `json:"volume_closed"`
	Profit         float64 `json:"profit"`
	Value          float64 `json:"value"`
	Storage        float64 `json:"storage"`
	Commission     float64 `json:"commission"`
	Fee            float64 `json:"fee"`
	RateProfit     float64 `json:"rate_profit"`
	RateMargin     float64 `json:"rate_margin"`
	ExpertId       int64   `json:"expert_id"`
	PositionId     int64   `json:"position_id"`
	Comment        string  `json:"comment"`
	ProfitRaw      float64 `json:"profit_raw"`
	PricePosition  float64 `json:"price_position"`
	TickValue      float64 `json:"tick_value"`
	TickSize       float64 `json:"tick_size"`
	Reason         int32   `json:"reason"`
	MarketBid      float64 `json:"market_bid"`
	MarketAsk      float64 `json:"market_ask"`
}

// Account is the money state of one login, mirroring hst.accounts.
type Account struct {
	Login          int64  `json:"login"`
	Group          string `json:"group"`
	Currency       string `json:"currency"`
	CurrencyDigits int32  `json:"currency_digits"`
	Leverage       int32  `json:"leverage"`

	Balance    float64 `json:"balance"`
	Credit     float64 `json:"credit"`
	Margin     float64 `json:"margin"`
	MarginFree float64 `json:"margin_free"`
	// MarginLevel is equity over margin as a percentage. Zero margin means no level.
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

	// Rights carries the account's own permission bits, so a disabled login is refused without
	// a second read.
	Rights int64 `json:"rights"`
}
