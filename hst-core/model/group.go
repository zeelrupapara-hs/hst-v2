package model

// Group is a set of accounts sharing settings, keyed by its path.
type Group struct {
	GroupId        int64
	Group          string
	Currency       string
	CurrencyDigits int32

	MarginMode       int32
	MarginFreeMode   int32
	MarginSOMode     int32
	MarginCall       float64
	MarginStopOut    float64
	MarginFlags      int32
	MarginFreeProfit int32

	TradeFlags         int32
	TradeInterestRate  float64
	TradeVirtualCredit float64

	LimitOrders         int32
	LimitPositions      int32
	LimitPositionsValue float64
	LimitSymbols        int32

	DemoDeposit  *float64
	DemoLeverage *int32
}

// Netting keeps one position per symbol; hedging lets many sit side by side.
type MarginMode int32

const (
	MarginRetailNetting MarginMode = 0
	MarginExchange      MarginMode = 1
	MarginRetailHedging MarginMode = 2
)

func (m MarginMode) Netting() bool { return m != MarginRetailHedging }

func (m MarginMode) Hedging() bool { return m == MarginRetailHedging }

// How the stop out level is read.
type StopOutMode int32

const (
	StopOutPercent StopOutMode = 0
	StopOutMoney   StopOutMode = 1
)

// GroupSymbol is what a group changes about one instrument; nil means it did not override.
type GroupSymbol struct {
	SymbolId    int64
	GroupId     int64
	Path        string
	ConfigIndex int32

	TradeMode  *int32
	ExecMode   *int32
	FillFlags  *int32
	ExpirFlags *int32

	SpreadDiff  *int32
	StopsLevel  *int32
	FreezeLevel *int32

	VolumeMin   *int64
	VolumeMax   *int64
	VolumeStep  *int64
	VolumeLimit *int64

	MarginInitial     *float64
	MarginMaintenance *float64
	MarginHedged      *float64
	MarginFlags       *int32

	SwapMode  *int32
	SwapLong  *float64
	SwapShort *float64
	SwapFlags *int32

	IECheckMode  *int32
	IETimeout    *int32
	IESlipProfit *int32
	IESlipLosing *int32
	IEVolumeMax  *int64
	IEFlags      *int32

	RETimeout *int32
	REFlags   *int32

	OrderFlags       *int32
	PermissionsFlags *int32
}
