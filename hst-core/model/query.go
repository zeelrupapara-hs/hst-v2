package model

type QueryRequest struct {
	Login int64     `json:"login"`
	What  QueryWhat `json:"what"`
}

type QueryResult struct {
	Login     int64        `json:"login"`
	Found     bool         `json:"found"`
	Account   *Account     `json:"account,omitempty"`
	Positions []Position   `json:"positions,omitempty"`
	Orders    []Order      `json:"orders,omitempty"`
	Symbols   []SymbolInfo `json:"symbols,omitempty"`
}

// SymbolInfo is one instrument under the group's resolved rules, not the instrument's own, plus its live price.
type SymbolInfo struct {
	Symbol          string `json:"symbol"`
	Path            string `json:"path"`
	Description     string `json:"description"`
	International   string `json:"international"`
	ISIN            string `json:"isin"`
	Category        string `json:"category"`
	Exchange        string `json:"exchange"`
	CFI             string `json:"cfi"`
	Industry        int32  `json:"industry"`
	Country         string `json:"country"`
	Basis           string `json:"basis"`
	Source          string `json:"source"`
	Page            string `json:"page"`
	ColorBackground int64  `json:"color_background"`

	Digits        int32   `json:"digits"`
	Point         float64 `json:"point"`
	ContractSize  float64 `json:"contract_size"`
	TickValue     float64 `json:"tick_value"`
	TickSize      float64 `json:"tick_size"`
	CalcMode      int32   `json:"calc_mode"`
	TradeMode     int32   `json:"trade_mode"`
	ExecMode      int32   `json:"exec_mode"`
	FillFlags     int32   `json:"fill_flags"`
	ExpirFlags    int32   `json:"expir_flags"`
	OrderFlags    int32   `json:"order_flags"`
	Sector        int32   `json:"sector"`
	GtcMode       int32   `json:"gtc_mode"`
	TickChartMode int32   `json:"tick_chart_mode"`

	// the volume limits are lots, already folded up from the extended units the engine counts in
	VolumeMin   float64 `json:"volume_min"`
	VolumeMax   float64 `json:"volume_max"`
	VolumeStep  float64 `json:"volume_step"`
	VolumeLimit float64 `json:"volume_limit"`

	// how far from the market an SL, TP or pending must sit, and how close it may no longer be moved
	StopsLevel  int32 `json:"stops_level"`
	FreezeLevel int32 `json:"freeze_level"`

	CurrencyBase   string `json:"currency_base"`
	CurrencyProfit string `json:"currency_profit"`
	CurrencyMargin string `json:"currency_margin"`

	MarginInitial     float64 `json:"margin_initial"`
	MarginMaintenance float64 `json:"margin_maintenance"`
	MarginHedged      float64 `json:"margin_hedged"`

	// the buy and sell market-order rates, which is what a terminal shows against a symbol
	MarginRateInitialBuy      float64 `json:"margin_rate_initial_buy"`
	MarginRateInitialSell     float64 `json:"margin_rate_initial_sell"`
	MarginRateMaintenanceBuy  float64 `json:"margin_rate_maintenance_buy"`
	MarginRateMaintenanceSell float64 `json:"margin_rate_maintenance_sell"`

	MarginSpec *SymbolMarginSpec `json:"margin_spec,omitempty"`

	SwapMode  int32   `json:"swap_mode"`
	SwapLong  float64 `json:"swap_long"`
	SwapShort float64 `json:"swap_short"`

	SwapRateSunday    float64 `json:"swap_rate_sunday"`
	SwapRateMonday    float64 `json:"swap_rate_monday"`
	SwapRateTuesday   float64 `json:"swap_rate_tuesday"`
	SwapRateWednesday float64 `json:"swap_rate_wednesday"`
	SwapRateThursday  float64 `json:"swap_rate_thursday"`
	SwapRateFriday    float64 `json:"swap_rate_friday"`
	SwapRateSaturday  float64 `json:"swap_rate_saturday"`
	SwapYearDay       int32   `json:"swap_year_day"`
	SwapFlags         int32   `json:"swap_flags"`

	// the live quote, zero when the instrument has not printed since the engine started
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	Last     float64 `json:"last"`
	Time     int64   `json:"time"`
	Gap      bool    `json:"gap"`
	HasQuote bool    `json:"has_quote"`

	// the session bar and the move off its close, zero when the feed carries no daily close
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

// SymbolMarginSpec is the floating leverage ladder that applies to this symbol for the group.
type SymbolMarginSpec struct {
	Floating  bool               `json:"floating"`
	RangeMode int32              `json:"range_mode"`
	RulePath  string             `json:"rule_path"`
	Tiers     []SymbolMarginTier `json:"tiers,omitempty"`
}

// SymbolMarginTier is one band of a floating margin rule; RangeTo 0 means infinity.
type SymbolMarginTier struct {
	RangeFrom             float64 `json:"range_from"`
	RangeTo               float64 `json:"range_to"`
	MarginRateInitial     float64 `json:"margin_rate_initial"`
	MarginRateMaintenance float64 `json:"margin_rate_maintenance"`
}

type QueryWhat int32

const (
	QueryWhat_account   QueryWhat = 1
	QueryWhat_positions QueryWhat = 2
	QueryWhat_orders    QueryWhat = 3
	QueryWhat_state     QueryWhat = 4
	QueryWhat_symbols   QueryWhat = 5
)

func (m TradeMode) AllowsBuy() bool { return m == TradeMode_full || m == TradeMode_long_only }

func (m TradeMode) AllowsSell() bool { return m == TradeMode_full || m == TradeMode_short_only }

func (m TradeMode) CloseOnly() bool { return m == TradeMode_close_only }

func (m MarginMode) Hedging() bool { return m == MarginMode_retail_hedging }

// A delay or a cleared level lets the request carry on; everything else settles it.
func (a RouteAction) Terminal() bool {
	switch a {
	case RouteAction_delay_time, RouteAction_delay_tick, RouteAction_clear_tp, RouteAction_clear_sl, RouteAction_clear_sltp:
		return false
	}
	return true
}

// ToDealer reports whether the action hands the request to the dealing desk.
func (a RouteAction) ToDealer() bool {
	return a == RouteAction_dealer || a == RouteAction_dealer_online
}

func (a RouteAction) Executes() bool {
	return a == RouteAction_confirm_client || a == RouteAction_confirm_market
}
