package model

// Position enums and record.

// PositionAction is the side the position holds. A position is only ever buy or sell.
type PositionAction int32

const (
	PositionAction_buy  PositionAction = 0
	PositionAction_sell PositionAction = 1
)

var PositionAction_name = map[int32]string{0: "buy", 1: "sell"}

// PositionReason is what opened the position. Same vocabulary as the order that caused it.
type PositionReason int32

const (
	PositionReason_client PositionReason = 0
	PositionReason_expert PositionReason = 1
	PositionReason_dealer PositionReason = 2
	PositionReason_signal PositionReason = 10
	PositionReason_mobile PositionReason = 16
	PositionReason_web    PositionReason = 17
)

// Position is one row of hst.positions. PriceOpen is the weighted average of what opened it.
type Position struct {
	PositionId       int64   `json:"position_id"`
	ExternalId       string  `json:"external_id"`
	Login            int64   `json:"login"`
	Dealer           int64   `json:"dealer"`
	Symbol           string  `json:"symbol"`
	Action           int32   `json:"action"`
	Digits           int32   `json:"digits"`
	DigitsCurrency   int32   `json:"digits_currency"`
	Reason           int32   `json:"reason"`
	ContractSize     float64 `json:"contract_size"`
	TimeCreate       int64   `json:"time_create"`
	TimeUpdate       int64   `json:"time_update"`
	PriceOpen        float64 `json:"price_open"`
	PriceCurrent     float64 `json:"price_current"`
	PriceSL          float64 `json:"price_sl"`
	PriceTP          float64 `json:"price_tp"`
	Volume           int64   `json:"volume"`
	VolumeExt        int64   `json:"volume_ext"`
	Profit           float64 `json:"profit"`
	Storage          float64 `json:"storage"`
	RateProfit       float64 `json:"rate_profit"`
	RateMargin       float64 `json:"rate_margin"`
	ExpertId         int64   `json:"expert_id"`
	ExpertPositionId int64   `json:"expert_position_id"`
	Comment          string  `json:"comment"`
	ActivationMode   int32   `json:"activation_mode"`
	ActivationTime   int64   `json:"activation_time"`
	ActivationPrice  float64 `json:"activation_price"`
	ActivationFlags  int32   `json:"activation_flags"`
	ApiData          string  `json:"api_data"`
	DateCreated      int64   `json:"date_created"`
	DateModified     int64   `json:"date_modified"`
}

// IsBuy reports whether the position is long.
func (p *Position) IsBuy() bool { return PositionAction(p.Action) == PositionAction_buy }

// VolumeLots is the position volume as a decimal number of lots.
func (p *Position) VolumeLots() float64 { return VolumeToLots(p.Volume) }
