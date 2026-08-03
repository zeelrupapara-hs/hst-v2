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
