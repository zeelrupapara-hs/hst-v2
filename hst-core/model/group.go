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

	SwapMode    *int32
	SwapRate    [7]*float64
	SwapYearDay *int32
	SwapLong    *float64
	SwapShort   *float64
	SwapFlags   *int32

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

// What the group lets its accounts do, from hst.groups.trade_flags.
const (
	TradeFlagSwaps                int32 = 0x0001
	TradeFlagTrailing             int32 = 0x0002
	TradeFlagExperts              int32 = 0x0004
	TradeFlagExpiration           int32 = 0x0008
	TradeFlagSignalsAll           int32 = 0x0010
	TradeFlagSignalsOwn           int32 = 0x0020
	TradeFlagSOCompensation       int32 = 0x0040
	TradeFlagSOFullyHedged        int32 = 0x0080
	TradeFlagFIFOClose            int32 = 0x0100
	TradeFlagHedgeProhibit        int32 = 0x0200
	TradeFlagDealCost             int32 = 0x0400
	TradeFlagSOCompensationCredit int32 = 0x0800
)

// How much of the floating result counts toward free margin, from hst.groups.margin_free_mode.
type FreeMarginMode int32

const (
	FreeMarginNotUsePL FreeMarginMode = 0
	FreeMarginUsePL    FreeMarginMode = 1
	FreeMarginProfit   FreeMarginMode = 2
	FreeMarginLoss     FreeMarginMode = 3
)

// Counts reports how much of a floating result the mode admits into free margin.
func (m FreeMarginMode) Counts(floating float64) float64 {
	switch m {
	case FreeMarginUsePL:
		return floating
	case FreeMarginProfit:
		if floating > 0 {
			return floating
		}
	case FreeMarginLoss:
		if floating < 0 {
			return floating
		}
	}

	return 0
}

// How a day's realised profit is treated, from hst.groups.margin_free_profit_mode.
type FreeMarginProfitMode int32

const (
	FreeMarginDayProfitAndLoss FreeMarginProfitMode = 0
	FreeMarginDayProfitLoss    FreeMarginProfitMode = 1
)

// The group's own margin flags, from hst.groups.margin_flags.
const GroupMarginFlagClearAccumulated int32 = 1

// What the group permits, from hst.groups.permission_flags.
const (
	PermissionEnableConnection int32 = 0x0002
	PermissionNotifyDeals      int32 = 0x0040
	PermissionNotifyOrders     int32 = 0x0080
	PermissionNotifyBalances   int32 = 0x0100
)

// Which order types and levels an instrument offers, from order_flags.
const (
	OrderFlagMarket    int32 = 1
	OrderFlagLimit     int32 = 2
	OrderFlagStop      int32 = 4
	OrderFlagStopLimit int32 = 8
	OrderFlagSL        int32 = 16
	OrderFlagTP        int32 = 32
	OrderFlagCloseBy   int32 = 64
)

// OrderFlagFor is the bit an order type needs to be allowed.
func OrderFlagFor(t OrderType) int32 {
	switch t {
	case OrderBuy, OrderSell:
		return OrderFlagMarket
	case OrderBuyLimit, OrderSellLimit:
		return OrderFlagLimit
	case OrderBuyStop, OrderSellStop:
		return OrderFlagStop
	case OrderBuyStopLimit, OrderSellStopLimit:
		return OrderFlagStopLimit
	case OrderCloseBy:
		return OrderFlagCloseBy
	}
	return 0
}
