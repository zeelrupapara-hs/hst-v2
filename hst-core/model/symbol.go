package model

// Symbol is an instrument, as far as the engine cares.
//
// hst.symbols has well over a hundred columns; only the ones a trade actually reads are loaded.
// Anything the engine never looks at stays in the database where the admin panel edits it.
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

	SwapMode   int32
	SwapLong   float64
	SwapShort  float64
	SwapFlags  int32
	SwapRate   [7]float64
	QuotesTime int32

	TimeStart      int64
	TimeExpiration int64
}

// How the margin for an instrument is worked out. From symbol_enum.htm.
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

// What a group may do with an instrument. From symbol_enum.htm.
type TradeMode int32

const (
	TradeDisabled  TradeMode = 0
	TradeLongOnly  TradeMode = 1
	TradeShortOnly TradeMode = 2
	TradeCloseOnly TradeMode = 3
	TradeFull      TradeMode = 4
)

// How a request is turned into a fill. From group_symbols_execution.htm.
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

// AllowsBuy reports whether the instrument may be bought.
func (m TradeMode) AllowsBuy() bool { return m == TradeFull || m == TradeLongOnly }

// AllowsSell reports whether the instrument may be sold.
func (m TradeMode) AllowsSell() bool { return m == TradeFull || m == TradeShortOnly }

// CloseOnly reports whether existing positions may be closed but nothing new opened.
func (m TradeMode) CloseOnly() bool { return m == TradeCloseOnly }
