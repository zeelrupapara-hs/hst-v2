package model

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

// VolumeUnit is one legacy volume unit: 1/10000 of a lot.
const VolumeUnit = 10000.0

// VolumeUnitExt is one extended volume unit: 1/100000000 of a lot. The engine counts in these,
// so a request carries them and the coarse columns are only ever read for display.
const VolumeUnitExt = 100000000.0

// ExtPerUnit is how many extended units one legacy unit is worth.
const ExtPerUnit = int64(VolumeUnitExt / VolumeUnit)

// ExtendedVolume picks whichever of the two columns a row actually carries.
func ExtendedVolume(volume, ext int64) int64 {
	if ext > 0 {
		return ext
	}

	return volume * ExtPerUnit
}

// ExtToLots is an extended volume as a decimal number of lots.
func ExtToLots(v int64) float64 { return float64(v) / VolumeUnitExt }

// VolumeToLots converts integer volume into lots.
func VolumeToLots(v int64) float64 { return float64(v) / VolumeUnit }

// LotsToVolume converts lots into integer volume.
func LotsToVolume(lots float64) int64 { return int64(lots*VolumeUnitExt + 0.5) }
