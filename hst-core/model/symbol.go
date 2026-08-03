package model

// Symbol is an instrument, with only the columns a trade actually reads.
type Symbol struct {
	SymbolId     int64
	Symbol       string
	Path         string
	Description  string
	Digits       int32
	Point        float64
	CalcMode     int32
	TradeMode    int32
	ExecMode     int32
	FillFlags    int32
	ExpirFlags   int32
	ContractSize float64
	TickValue    float64
	TickSize     float64
	Spread       int32
	SpreadDiff   int32
	StopsLevel   int32
	FreezeLevel  int32

	CurrencyBase   string
	CurrencyProfit string
	CurrencyMargin string

	VolumeMin   int64
	VolumeMax   int64
	VolumeStep  int64
	VolumeLimit int64

	VolumeMinExt   int64
	VolumeMaxExt   int64
	VolumeStepExt  int64
	VolumeLimitExt int64

	MarginInitial     float64
	MarginMaintenance float64
	MarginHedged      float64
	MarginFlags       int32

	// MarginRates multiply the margin a formula produced, one per order type.
	MarginRateInitial     MarginRates
	MarginRateMaintenance MarginRates

	SwapMode    int32
	SwapLong    float64
	SwapShort   float64
	SwapFlags   int32
	SwapRate    [7]float64
	SwapYearDay int32
	QuotesTime  int32
	OrderFlags  int32

	TimeStart      int64
	TimeExpiration int64
}

// Which filling policies a symbol allows, as a set of flags.
const (
	FillFlagFOK    = 0x01
	FillFlagIOC    = 0x02
	FillFlagReturn = 0x04
	FillFlagBOC    = 0x08
)

// Which expiry types a symbol allows, as a set of flags.
const (
	ExpirFlagGTC          = 0x01
	ExpirFlagDay          = 0x02
	ExpirFlagSpecified    = 0x04
	ExpirFlagSpecifiedDay = 0x08
)

// How margin is checked and shared, from margin_flags on the instrument or the group's override.
const (
	MarginFlagCheckProcess  int32 = 1
	MarginFlagCheckSLTP     int32 = 2
	MarginFlagHedgeLargeLeg int32 = 4
	MarginFlagExcludePL     int32 = 8
	MarginFlagRecalcRates   int32 = 16
)

// SwapFlagConsiderHolidays doubles the swap the day before a holiday and charges none on it.
const SwapFlagConsiderHolidays int32 = 1

// InstantCheckNormal is the only instant execution check mode the platform defines.
const InstantCheckNormal int32 = 0

// InstantFlagFastConfirmation lets an accepted requote execute without going back to the
// dealer, when the price is inside the deviation the client themselves allowed.
const InstantFlagFastConfirmation int32 = 1

// RequestFlagOrder makes the dealer confirm once more after the client accepts their price.
const RequestFlagOrder int32 = 1

// MarginRates is a multiplier per order type, in OrderType order.
type MarginRates [8]float64

// For is the rate an order type carries. A maintenance rate of zero means the initial one stands.
func (m MarginRates) For(t OrderType) float64 {
	if int(t) < 0 || int(t) >= len(m) {
		return 1
	}

	return m[t]
}
