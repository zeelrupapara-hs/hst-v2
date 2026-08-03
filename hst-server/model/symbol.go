package model

// Symbol-domain enums.

var (
	CalcMode_name = map[int32]string{
		0: "forex", 1: "futures", 2: "cfd", 3: "cfdindex", 4: "cfdleverage",
		5: "forex_no_leverage", 32: "exch_stocks", 33: "exch_futures", 34: "exch_forts",
		35: "exch_options", 36: "exch_options_margin", 37: "exch_bonds", 64: "serv_collateral",
	}
	CalcMode_value = map[string]int32{
		"forex": 0, "futures": 1, "cfd": 2, "cfdindex": 3, "cfdleverage": 4,
		"forex_no_leverage": 5, "exch_stocks": 32, "exch_futures": 33, "exch_forts": 34,
		"exch_options": 35, "exch_options_margin": 36, "exch_bonds": 37, "serv_collateral": 64,
	}
)

var (
	TradeMode_name = map[int32]string{
		0: "disabled", 1: "longonly", 2: "shortonly", 3: "closeonly", 4: "full",
	}
	TradeMode_value = map[string]int32{
		"disabled": 0, "longonly": 1, "shortonly": 2, "closeonly": 3, "full": 4,
	}
)

var (
	ExecMode_name = map[int32]string{
		0: "request", 1: "instant", 2: "market", 3: "exchange",
	}
	ExecMode_value = map[string]int32{
		"request": 0, "instant": 1, "market": 2, "exchange": 3,
	}
)

type GTCMode int32

const (
	GTCMode_gtc            GTCMode = 0
	GTCMode_daily          GTCMode = 1
	GTCMode_daily_no_stops GTCMode = 2
)

var (
	GTCMode_name = map[int32]string{
		0: "gtc", 1: "daily", 2: "daily_no_stops",
	}
	GTCMode_value = map[string]int32{
		"gtc": 0, "daily": 1, "daily_no_stops": 2,
	}
)

type FillingFlags int32

const (
	FillingFlags_none FillingFlags = 0
	FillingFlags_fok  FillingFlags = 1
	FillingFlags_ioc  FillingFlags = 2
)

var (
	FillingFlags_name = map[int32]string{
		0: "none", 1: "fok", 2: "ioc",
	}
	FillingFlags_value = map[string]int32{
		"none": 0, "fok": 1, "ioc": 2,
	}
)

type ExpirationFlags int32

const (
	ExpirationFlags_none          ExpirationFlags = 0
	ExpirationFlags_gtc           ExpirationFlags = 1
	ExpirationFlags_day           ExpirationFlags = 2
	ExpirationFlags_specified     ExpirationFlags = 4
	ExpirationFlags_specified_day ExpirationFlags = 8
)

var (
	ExpirationFlags_name = map[int32]string{
		0: "none", 1: "gtc", 2: "day", 4: "specified", 8: "specified_day",
	}
	ExpirationFlags_value = map[string]int32{
		"none": 0, "gtc": 1, "day": 2, "specified": 4, "specified_day": 8,
	}
)

type OrderFlags int32

const (
	OrderFlags_none       OrderFlags = 0
	OrderFlags_market     OrderFlags = 1
	OrderFlags_limit      OrderFlags = 2
	OrderFlags_stop       OrderFlags = 4
	OrderFlags_stop_limit OrderFlags = 8
	OrderFlags_sl         OrderFlags = 16
	OrderFlags_tp         OrderFlags = 32
	OrderFlags_closeby    OrderFlags = 64

	OrderFlags_all OrderFlags = 127
)

var (
	OrderFlags_name = map[int32]string{
		0: "none", 1: "market", 2: "limit", 4: "stop", 8: "stop_limit",
		16: "sl", 32: "tp", 64: "closeby",
	}
	OrderFlags_value = map[string]int32{
		"none": 0, "market": 1, "limit": 2, "stop": 4, "stop_limit": 8,
		"sl": 16, "tp": 32, "closeby": 64,
	}
)

type SwapMode int32

const (
	SwapMode_disabled              SwapMode = 0
	SwapMode_by_points             SwapMode = 1
	SwapMode_by_symbol_currency    SwapMode = 2
	SwapMode_by_margin_currency    SwapMode = 3
	SwapMode_by_group_currency     SwapMode = 4
	SwapMode_by_interest_current   SwapMode = 5
	SwapMode_by_interest_open      SwapMode = 6
	SwapMode_reopen_by_close_price SwapMode = 7
	SwapMode_reopen_by_bid         SwapMode = 8
	SwapMode_by_profit_currency    SwapMode = 9
)

var (
	SwapMode_name = map[int32]string{
		0: "disabled", 1: "by_points", 2: "by_symbol_currency", 3: "by_margin_currency",
		4: "by_group_currency", 5: "by_interest_current", 6: "by_interest_open",
		7: "reopen_by_close_price", 8: "reopen_by_bid", 9: "by_profit_currency",
	}
	SwapMode_value = map[string]int32{
		"disabled": 0, "by_points": 1, "by_symbol_currency": 2, "by_margin_currency": 3,
		"by_group_currency": 4, "by_interest_current": 5, "by_interest_open": 6,
		"reopen_by_close_price": 7, "reopen_by_bid": 8, "by_profit_currency": 9,
	}
)

type SwapDays int32

const (
	SwapDays_sunday    SwapDays = 0
	SwapDays_monday    SwapDays = 1
	SwapDays_tuesday   SwapDays = 2
	SwapDays_wednesday SwapDays = 3
	SwapDays_thursday  SwapDays = 4
	SwapDays_friday    SwapDays = 5
	SwapDays_saturday  SwapDays = 6
	SwapDays_disabled  SwapDays = 7
)

var (
	SwapDays_name = map[int32]string{
		0: "sunday", 1: "monday", 2: "tuesday", 3: "wednesday",
		4: "thursday", 5: "friday", 6: "saturday", 7: "disabled",
	}
	SwapDays_value = map[string]int32{
		"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3,
		"thursday": 4, "friday": 5, "saturday": 6, "disabled": 7,
	}
)

type SwapFlags int32

const (
	SwapFlags_none              SwapFlags = 0
	SwapFlags_consider_holidays SwapFlags = 1
)

var (
	SwapFlags_name  = map[int32]string{0: "none", 1: "consider_holidays"}
	SwapFlags_value = map[string]int32{"none": 0, "consider_holidays": 1}
)

// SymbolMarginFlags is the symbol-level EnMarginFlags (not group-level).
type SymbolMarginFlags int32

const (
	SymbolMarginFlags_none            SymbolMarginFlags = 0
	SymbolMarginFlags_check_process   SymbolMarginFlags = 1
	SymbolMarginFlags_check_sltp      SymbolMarginFlags = 2
	SymbolMarginFlags_hedge_large_leg SymbolMarginFlags = 4
	SymbolMarginFlags_exclude_pl      SymbolMarginFlags = 8
	SymbolMarginFlags_recalc_rates    SymbolMarginFlags = 16

	SymbolMarginFlags_all SymbolMarginFlags = 31
)

var (
	SymbolMarginFlags_name = map[int32]string{
		0: "none", 1: "check_process", 2: "check_sltp", 4: "hedge_large_leg",
		8: "exclude_pl", 16: "recalc_rates",
	}
	SymbolMarginFlags_value = map[string]int32{
		"none": 0, "check_process": 1, "check_sltp": 2, "hedge_large_leg": 4,
		"exclude_pl": 8, "recalc_rates": 16,
	}
)

type TickFlags int32

const (
	TickFlags_none       TickFlags = 0
	TickFlags_realtime   TickFlags = 1
	TickFlags_collectraw TickFlags = 2
	TickFlags_feed_stats TickFlags = 4
)

var (
	TickFlags_name = map[int32]string{
		0: "none", 1: "realtime", 2: "collectraw", 4: "feed_stats",
	}
	TickFlags_value = map[string]int32{
		"none": 0, "realtime": 1, "collectraw": 2, "feed_stats": 4,
	}
)

type ChartMode int32

const (
	ChartMode_bid_price  ChartMode = 0
	ChartMode_last_price ChartMode = 1
	ChartMode_old        ChartMode = 255
)

var (
	ChartMode_name  = map[int32]string{0: "bid_price", 1: "last_price", 255: "old"}
	ChartMode_value = map[string]int32{"bid_price": 0, "last_price": 1, "old": 255}
)

type OptionMode int32

const (
	OptionMode_european_call OptionMode = 0
	OptionMode_european_put  OptionMode = 1
	OptionMode_american_call OptionMode = 2
	OptionMode_american_put  OptionMode = 3
)

var (
	OptionMode_name = map[int32]string{
		0: "european_call", 1: "european_put", 2: "american_call", 3: "american_put",
	}
	OptionMode_value = map[string]int32{
		"european_call": 0, "european_put": 1, "american_call": 2, "american_put": 3,
	}
)

type SpliceType int32

const (
	SpliceType_none       SpliceType = 0
	SpliceType_unadjusted SpliceType = 1
	SpliceType_adjusted   SpliceType = 2
)

var (
	SpliceType_name  = map[int32]string{0: "none", 1: "unadjusted", 2: "adjusted"}
	SpliceType_value = map[string]int32{"none": 0, "unadjusted": 1, "adjusted": 2}
)

type SpliceTimeType int32

const (
	SpliceTimeType_expiration SpliceTimeType = 0
)

var (
	SpliceTimeType_name  = map[int32]string{0: "expiration"}
	SpliceTimeType_value = map[string]int32{"expiration": 0}
)

type InstantMode int32

const (
	InstantMode_check_normal InstantMode = 0
)

var (
	InstantMode_name  = map[int32]string{0: "check_normal"}
	InstantMode_value = map[string]int32{"check_normal": 0}
)

// InstantFlags is EnInstantFlags: what the instant execution mode may do.
type InstantFlags int32

const (
	InstantFlags_none              InstantFlags = 0
	InstantFlags_fast_confirmation InstantFlags = 1
)

var (
	InstantFlags_name  = map[int32]string{0: "none", 1: "fast_confirmation"}
	InstantFlags_value = map[string]int32{"none": 0, "fast_confirmation": 1}
)

type RequestFlags int32

const (
	RequestFlags_none  RequestFlags = 0
	RequestFlags_order RequestFlags = 1
)

var (
	RequestFlags_name  = map[int32]string{0: "none", 1: "order"}
	RequestFlags_value = map[string]int32{"none": 0, "order": 1}
)

// SymbolTradeFlags is the symbol-level EnTradeFlags (not group-level).
type SymbolTradeFlags int32

const (
	SymbolTradeFlags_none             SymbolTradeFlags = 0
	SymbolTradeFlags_profit_by_market SymbolTradeFlags = 1
	SymbolTradeFlags_allow_signals    SymbolTradeFlags = 2
)

var (
	SymbolTradeFlags_name = map[int32]string{
		0: "none", 1: "profit_by_market", 2: "allow_signals",
	}
	SymbolTradeFlags_value = map[string]int32{
		"none": 0, "profit_by_market": 1, "allow_signals": 2,
	}
)

type SymbolSector int32

const (
	SymbolSector_undefined              SymbolSector = 0
	SymbolSector_basic_materials        SymbolSector = 1
	SymbolSector_communication_services SymbolSector = 2
	SymbolSector_consumer_cyclical      SymbolSector = 3
	SymbolSector_consumer_defensive     SymbolSector = 4
	SymbolSector_energy                 SymbolSector = 5
	SymbolSector_financial              SymbolSector = 6
	SymbolSector_healthcare             SymbolSector = 7
	SymbolSector_industrials            SymbolSector = 8
	SymbolSector_real_estate            SymbolSector = 9
	SymbolSector_technology             SymbolSector = 10
	SymbolSector_utilities              SymbolSector = 11
	SymbolSector_currency               SymbolSector = 12
	SymbolSector_currency_crypto        SymbolSector = 13
	SymbolSector_indexes                SymbolSector = 14
	SymbolSector_commodities            SymbolSector = 15
)

var (
	SymbolSector_name = map[int32]string{
		0: "undefined", 1: "basic_materials", 2: "communication_services",
		3: "consumer_cyclical", 4: "consumer_defensive", 5: "energy", 6: "financial",
		7: "healthcare", 8: "industrials", 9: "real_estate", 10: "technology",
		11: "utilities", 12: "currency", 13: "currency_crypto", 14: "indexes", 15: "commodities",
	}
	SymbolSector_value = map[string]int32{
		"undefined": 0, "basic_materials": 1, "communication_services": 2,
		"consumer_cyclical": 3, "consumer_defensive": 4, "energy": 5, "financial": 6,
		"healthcare": 7, "industrials": 8, "real_estate": 9, "technology": 10,
		"utilities": 11, "currency": 12, "currency_crypto": 13, "indexes": 14, "commodities": 15,
	}
)

// SymbolIndustry stores EnIndustries.
type SymbolIndustry int32

const (
	SymbolIndustry_undefined SymbolIndustry = 0
)

type SymbolSessionType int32

const (
	SymbolSessionType_quote SymbolSessionType = 0
	SymbolSessionType_trade SymbolSessionType = 1
)

var (
	SymbolSessionType_name  = map[int32]string{0: "quote", 1: "trade"}
	SymbolSessionType_value = map[string]int32{"quote": 0, "trade": 1}
)

// Symbol is the server-wide instrument master.
type Symbol struct {
	SymbolId                       int64             `db:"symbol_id" json:"symbol_id"`
	Symbol                         string            `db:"symbol" json:"symbol"`
	Path                           string            `db:"path" json:"path"`
	Isin                           string            `db:"isin" json:"isin"`
	Description                    string            `db:"description" json:"description"`
	International                  string            `db:"international" json:"international"`
	Category                       string            `db:"category" json:"category"`
	Exchange                       string            `db:"exchange" json:"exchange"`
	Cfi                            string            `db:"cfi" json:"cfi"`
	Sector                         SymbolSector      `db:"sector" json:"sector"`
	Industry                       SymbolIndustry    `db:"industry" json:"industry"`
	Country                        string            `db:"country" json:"country"`
	Basis                          string            `db:"basis" json:"basis"`
	Source                         string            `db:"source" json:"source"`
	Page                           string            `db:"page" json:"page"`
	CurrencyBase                   string            `db:"currency_base" json:"currency_base"`
	CurrencyBaseDigits             int32             `db:"currency_base_digits" json:"currency_base_digits"`
	CurrencyProfit                 string            `db:"currency_profit" json:"currency_profit"`
	CurrencyProfitDigits           int32             `db:"currency_profit_digits" json:"currency_profit_digits"`
	CurrencyMargin                 string            `db:"currency_margin" json:"currency_margin"`
	CurrencyMarginDigits           int32             `db:"currency_margin_digits" json:"currency_margin_digits"`
	Color                          int64             `db:"color" json:"color"`
	ColorBackground                int64             `db:"color_background" json:"color_background"`
	Digits                         int32             `db:"digits" json:"digits"`
	Point                          float64           `db:"point" json:"point"`
	Multiply                       float64           `db:"multiply" json:"multiply"`
	TickFlags                      TickFlags         `db:"tick_flags" json:"tick_flags"`
	TickBookDepth                  int32             `db:"tick_book_depth" json:"tick_book_depth"`
	FilterSoft                     int32             `db:"filter_soft" json:"filter_soft"`
	FilterSoftTicks                int32             `db:"filter_soft_ticks" json:"filter_soft_ticks"`
	FilterHard                     int32             `db:"filter_hard" json:"filter_hard"`
	FilterHardTicks                int32             `db:"filter_hard_ticks" json:"filter_hard_ticks"`
	FilterDiscard                  int32             `db:"filter_discard" json:"filter_discard"`
	FilterSpreadMax                int32             `db:"filter_spread_max" json:"filter_spread_max"`
	FilterSpreadMin                int32             `db:"filter_spread_min" json:"filter_spread_min"`
	SubscriptionsDelay             int32             `db:"subscriptions_delay" json:"subscriptions_delay"`
	TradeMode                      TradeMode         `db:"trade_mode" json:"trade_mode"`
	CalcMode                       CalcMode          `db:"calc_mode" json:"calc_mode"`
	ExecMode                       ExecMode          `db:"exec_mode" json:"exec_mode"`
	GtcMode                        GTCMode           `db:"gtc_mode" json:"gtc_mode"`
	FillFlags                      FillingFlags      `db:"fill_flags" json:"fill_flags"`
	ExpirFlags                     ExpirationFlags   `db:"expir_flags" json:"expir_flags"`
	Spread                         int32             `db:"spread" json:"spread"`
	SpreadBalance                  int32             `db:"spread_balance" json:"spread_balance"`
	SpreadDiff                     int32             `db:"spread_diff" json:"spread_diff"`
	SpreadDiffBalance              int32             `db:"spread_diff_balance" json:"spread_diff_balance"`
	TickValue                      float64           `db:"tick_value" json:"tick_value"`
	TickSize                       float64           `db:"tick_size" json:"tick_size"`
	ContractSize                   float64           `db:"contract_size" json:"contract_size"`
	StopsLevel                     int32             `db:"stops_level" json:"stops_level"`
	FreezeLevel                    int32             `db:"freeze_level" json:"freeze_level"`
	QuotesTimeout                  int32             `db:"quotes_timeout" json:"quotes_timeout"`
	VolumeMin                      int64             `db:"volume_min" json:"volume_min"`
	VolumeMinExt                   int64             `db:"volume_min_ext" json:"volume_min_ext"`
	VolumeMax                      int64             `db:"volume_max" json:"volume_max"`
	VolumeMaxExt                   int64             `db:"volume_max_ext" json:"volume_max_ext"`
	VolumeStep                     int64             `db:"volume_step" json:"volume_step"`
	VolumeStepExt                  int64             `db:"volume_step_ext" json:"volume_step_ext"`
	VolumeLimit                    int64             `db:"volume_limit" json:"volume_limit"`
	VolumeLimitExt                 int64             `db:"volume_limit_ext" json:"volume_limit_ext"`
	MarginFlags                    SymbolMarginFlags `db:"margin_flags" json:"margin_flags"`
	MarginInitial                  float64           `db:"margin_initial" json:"margin_initial"`
	MarginMaintenance              float64           `db:"margin_maintenance" json:"margin_maintenance"`
	MarginInitialBuy               float64           `db:"margin_initial_buy" json:"margin_initial_buy"`
	MarginInitialSell              float64           `db:"margin_initial_sell" json:"margin_initial_sell"`
	MarginInitialBuyLimit          float64           `db:"margin_initial_buy_limit" json:"margin_initial_buy_limit"`
	MarginInitialSellLimit         float64           `db:"margin_initial_sell_limit" json:"margin_initial_sell_limit"`
	MarginInitialBuyStop           float64           `db:"margin_initial_buy_stop" json:"margin_initial_buy_stop"`
	MarginInitialSellStop          float64           `db:"margin_initial_sell_stop" json:"margin_initial_sell_stop"`
	MarginInitialBuyStopLimit      float64           `db:"margin_initial_buy_stop_limit" json:"margin_initial_buy_stop_limit"`
	MarginInitialSellStopLimit     float64           `db:"margin_initial_sell_stop_limit" json:"margin_initial_sell_stop_limit"`
	MarginMaintenanceBuy           float64           `db:"margin_maintenance_buy" json:"margin_maintenance_buy"`
	MarginMaintenanceSell          float64           `db:"margin_maintenance_sell" json:"margin_maintenance_sell"`
	MarginMaintenanceBuyLimit      float64           `db:"margin_maintenance_buy_limit" json:"margin_maintenance_buy_limit"`
	MarginMaintenanceSellLimit     float64           `db:"margin_maintenance_sell_limit" json:"margin_maintenance_sell_limit"`
	MarginMaintenanceBuyStop       float64           `db:"margin_maintenance_buy_stop" json:"margin_maintenance_buy_stop"`
	MarginMaintenanceSellStop      float64           `db:"margin_maintenance_sell_stop" json:"margin_maintenance_sell_stop"`
	MarginMaintenanceBuyStopLimit  float64           `db:"margin_maintenance_buy_stop_limit" json:"margin_maintenance_buy_stop_limit"`
	MarginMaintenanceSellStopLimit float64           `db:"margin_maintenance_sell_stop_limit" json:"margin_maintenance_sell_stop_limit"`
	MarginHedged                   float64           `db:"margin_hedged" json:"margin_hedged"`
	SwapMode                       SwapMode          `db:"swap_mode" json:"swap_mode"`
	SwapLong                       float64           `db:"swap_long" json:"swap_long"`
	SwapShort                      float64           `db:"swap_short" json:"swap_short"`
	SwapYearDay                    SwapDays          `db:"swap_year_day" json:"swap_year_day"`
	SwapFlags                      SwapFlags         `db:"swap_flags" json:"swap_flags"`
	SwapRateSunday                 float64           `db:"swap_rate_sunday" json:"swap_rate_sunday"`
	SwapRateMonday                 float64           `db:"swap_rate_monday" json:"swap_rate_monday"`
	SwapRateTuesday                float64           `db:"swap_rate_tuesday" json:"swap_rate_tuesday"`
	SwapRateWednesday              float64           `db:"swap_rate_wednesday" json:"swap_rate_wednesday"`
	SwapRateThursday               float64           `db:"swap_rate_thursday" json:"swap_rate_thursday"`
	SwapRateFriday                 float64           `db:"swap_rate_friday" json:"swap_rate_friday"`
	SwapRateSaturday               float64           `db:"swap_rate_saturday" json:"swap_rate_saturday"`
	TimeStart                      int64             `db:"time_start" json:"time_start"`
	TimeExpiration                 int64             `db:"time_expiration" json:"time_expiration"`
	ReFlags                        RequestFlags      `db:"re_flags" json:"re_flags"`
	ReTimeout                      int32             `db:"re_timeout" json:"re_timeout"`
	IeCheckMode                    InstantMode       `db:"ie_check_mode" json:"ie_check_mode"`
	IeTimeout                      int32             `db:"ie_timeout" json:"ie_timeout"`
	IeSlipProfit                   int32             `db:"ie_slip_profit" json:"ie_slip_profit"`
	IeSlipLosing                   int32             `db:"ie_slip_losing" json:"ie_slip_losing"`
	IeVolumeMax                    int64             `db:"ie_volume_max" json:"ie_volume_max"`
	IeVolumeMaxExt                 int64             `db:"ie_volume_max_ext" json:"ie_volume_max_ext"`
	PriceSettle                    float64           `db:"price_settle" json:"price_settle"`
	PriceLimitMax                  float64           `db:"price_limit_max" json:"price_limit_max"`
	PriceLimitMin                  float64           `db:"price_limit_min" json:"price_limit_min"`
	TradeFlags                     SymbolTradeFlags  `db:"trade_flags" json:"trade_flags"`
	OrderFlags                     OrderFlags        `db:"order_flags" json:"order_flags"`
	MarginRateLiquidity            float64           `db:"margin_rate_liquidity" json:"margin_rate_liquidity"`
	MarginRateCurrency             float64           `db:"margin_rate_currency" json:"margin_rate_currency"`
	FaceValue                      float64           `db:"face_value" json:"face_value"`
	AccruedInterest                float64           `db:"accrued_interest" json:"accrued_interest"`
	SpliceType                     SpliceType        `db:"splice_type" json:"splice_type"`
	SpliceTimeType                 SpliceTimeType    `db:"splice_time_type" json:"splice_time_type"`
	SpliceTimeDays                 int32             `db:"splice_time_days" json:"splice_time_days"`
	OptionMode                     OptionMode        `db:"option_mode" json:"option_mode"`
	PriceStrike                    float64           `db:"price_strike" json:"price_strike"`
	FilterGap                      int32             `db:"filter_gap" json:"filter_gap"`
	FilterGapTicks                 int32             `db:"filter_gap_ticks" json:"filter_gap_ticks"`
	TickChartMode                  ChartMode         `db:"tick_chart_mode" json:"tick_chart_mode"`
	DateCreated                    int64             `db:"date_created" json:"date_created"`
	DateModified                   int64             `db:"date_modified" json:"date_modified"`
}

func (Symbol) TableName() string { return "hst.symbols" }

// SymbolSession is one quote or trade window for a symbol weekday.
type SymbolSession struct {
	SessionId int64             `db:"session_id" json:"session_id"`
	SymbolId  int64             `db:"symbol_id" json:"symbol_id"`
	Type      SymbolSessionType `db:"type" json:"type"`
	Day       int16             `db:"day" json:"day"`
	Open      int32             `db:"open" json:"open"`
	Close     int32             `db:"close" json:"close"`
}

func (SymbolSession) TableName() string { return "hst.symbols_sessions" }
