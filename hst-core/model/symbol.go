package model

// Symbol is an instrument, with only the columns a trade actually reads.
type Symbol struct {
	SymbolId     int64
	Symbol       string
	Path         string
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

	MarginInitial     float64
	MarginMaintenance float64
	MarginHedged      float64
	MarginFlags       int32

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

// How the margin for an instrument is worked out.
type CalcMode int32

const (
	CalcForex             CalcMode = 0
	CalcFutures           CalcMode = 1
	CalcCFD               CalcMode = 2
	CalcCFDIndex          CalcMode = 3
	CalcCFDLeverage       CalcMode = 4
	CalcForexNoLeverage   CalcMode = 5
	CalcExchStocks        CalcMode = 32
	CalcExchFutures       CalcMode = 33
	CalcExchFORTS         CalcMode = 34
	CalcExchOptions       CalcMode = 35
	CalcExchOptionsMargin CalcMode = 36
)

// What a group may do with an instrument.
type TradeMode int32

const (
	TradeDisabled  TradeMode = 0
	TradeLongOnly  TradeMode = 1
	TradeShortOnly TradeMode = 2
	TradeCloseOnly TradeMode = 3
	TradeFull      TradeMode = 4
)

// How a request is turned into a fill.
type ExecMode int32

const (
	ExecRequest  ExecMode = 0
	ExecInstant  ExecMode = 1
	ExecMarket   ExecMode = 2
	ExecExchange ExecMode = 3
)

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

func (m TradeMode) AllowsBuy() bool { return m == TradeFull || m == TradeLongOnly }

func (m TradeMode) AllowsSell() bool { return m == TradeFull || m == TradeShortOnly }

func (m TradeMode) CloseOnly() bool { return m == TradeCloseOnly }

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
