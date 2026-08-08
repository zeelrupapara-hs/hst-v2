package model

import "strings"

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

	MarginLeverageId *int64
	// LeverageRules is the profile's tree, already attached at load; nil when the group has none.
	LeverageRules []*LeverageRule
}

// LeverageRule scopes a set of leverage tiers to a symbol path mask.
type LeverageRule struct {
	Path      string
	RangeMode int32
	Tiers     []LeverageTier
}

// LeverageTier is one level of a rule; RangeTo 0 on the last tier means infinity.
type LeverageTier struct {
	RangeTo               float64
	MarginRateInitial     float64
	MarginRateMaintenance float64
}

// How a rule reads the size a tier is chosen by, from hst.leverage_rules.range_mode.
const (
	RangeModeVolume          int32 = 0
	RangeModeVolumePerSymbol int32 = 1
)

// MarginRateFor is the floating margin rate for a size on one instrument, if the group has one.
//
// The first rule whose path covers the symbol wins, and within it the first tier the size falls
// into. Only the volume modes are answered: the notional ones need a price the caller has not got.
func (g *Group) MarginRateFor(symbol, path string, lots float64) (float64, bool) {
	for _, r := range g.LeverageRules {
		if r.RangeMode != RangeModeVolume && r.RangeMode != RangeModeVolumePerSymbol {
			continue
		}
		if !pathCovers(r.Path, symbol, path) {
			continue
		}

		for j := range r.Tiers {
			if t := &r.Tiers[j]; t.RangeTo <= 0 || lots <= t.RangeTo {
				return t.MarginRateInitial, true
			}
		}
	}

	return 0, false
}

// pathCovers reports whether a mask covers an instrument, by name or by its place in the tree.
func pathCovers(mask, symbol, path string) bool {
	switch {
	case mask == "" || mask == "*":
		return true
	case strings.HasSuffix(mask, "*"):
		prefix := strings.TrimSuffix(mask, "*")
		return strings.HasPrefix(path, prefix) || strings.HasPrefix(symbol, prefix)
	default:
		return mask == symbol || mask == path
	}
}

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

	SpreadDiff        *int32
	SpreadDiffBalance *int32
	StopsLevel        *int32
	FreezeLevel       *int32

	VolumeMin   *int64
	VolumeMax   *int64
	VolumeStep  *int64
	VolumeLimit *int64

	VolumeMinExt   *int64
	VolumeMaxExt   *int64
	VolumeStepExt  *int64
	VolumeLimitExt *int64

	MarginInitial     *float64
	MarginMaintenance *float64
	MarginHedged      *float64
	MarginFlags       *int32

	MarginRateInitial     [8]*float64
	MarginRateMaintenance [8]*float64

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
	case OrderType_buy, OrderType_sell:
		return OrderFlagMarket
	case OrderType_buy_limit, OrderType_sell_limit:
		return OrderFlagLimit
	case OrderType_buy_stop, OrderType_sell_stop:
		return OrderFlagStop
	case OrderType_buy_stop_limit, OrderType_sell_stop_limit:
		return OrderFlagStopLimit
	case OrderType_close_by:
		return OrderFlagCloseBy
	}
	return 0
}

// Netting keeps one position per symbol; hedging lets many sit side by side.
type MarginMode int32

const (
	MarginMode_retail_netting MarginMode = 0
	MarginMode_exchange       MarginMode = 1
	MarginMode_retail_hedging MarginMode = 2
)

// How the stop out level is read.
type StopOutMode int32

const (
	StopOutMode_percent StopOutMode = 0
	StopOutMode_money   StopOutMode = 1
)

// How much of the floating result counts toward free margin, from hst.groups.margin_free_mode.
type FreeMarginMode int32

const (
	FreeMarginMode_not_use_pl FreeMarginMode = 0
	FreeMarginMode_use_pl     FreeMarginMode = 1
	FreeMarginMode_profit     FreeMarginMode = 2
	FreeMarginMode_loss       FreeMarginMode = 3
)

// How a day's realised profit is treated, from hst.groups.margin_free_profit_mode.
type FreeMarginProfitMode int32

const (
	FreeMarginProfitMode_day_profit_and_loss FreeMarginProfitMode = 0
	FreeMarginProfitMode_day_profit_loss     FreeMarginProfitMode = 1
)
