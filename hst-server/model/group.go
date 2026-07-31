package model

// PermissionsFlags is a group permissions bitmask — combine with OR.
type PermissionsFlags int32

const (
	PermissionsFlags_none               PermissionsFlags = 0x00000000
	PermissionsFlags_cert_confirm       PermissionsFlags = 0x00000001
	PermissionsFlags_enable_connection  PermissionsFlags = 0x00000002
	PermissionsFlags_reset_password     PermissionsFlags = 0x00000004
	PermissionsFlags_forced_otp_usage   PermissionsFlags = 0x00000008
	PermissionsFlags_risk_warning       PermissionsFlags = 0x00000010
	PermissionsFlags_regulation_protect PermissionsFlags = 0x00000020
	PermissionsFlags_notify_deals       PermissionsFlags = 0x00000040
	PermissionsFlags_notify_orders      PermissionsFlags = 0x00000080
	PermissionsFlags_notify_balances    PermissionsFlags = 0x00000100
)

// Enum value maps for PermissionsFlags.
var (
	PermissionsFlags_name = map[int32]string{
		0x00000000: "none",
		0x00000001: "cert_confirm",
		0x00000002: "enable_connection",
		0x00000004: "reset_password",
		0x00000008: "forced_otp_usage",
		0x00000010: "risk_warning",
		0x00000020: "regulation_protect",
		0x00000040: "notify_deals",
		0x00000080: "notify_orders",
		0x00000100: "notify_balances",
	}
	PermissionsFlags_value = map[string]int32{
		"none":               0x00000000,
		"cert_confirm":       0x00000001,
		"enable_connection":  0x00000002,
		"reset_password":     0x00000004,
		"forced_otp_usage":   0x00000008,
		"risk_warning":       0x00000010,
		"regulation_protect": 0x00000020,
		"notify_deals":       0x00000040,
		"notify_orders":      0x00000080,
		"notify_balances":    0x00000100,
	}
)

// Has reports whether every bit in flag is set.
func (f PermissionsFlags) Has(flag PermissionsFlags) bool { return f&flag == flag }

// AuthMode is the client auth mode for a group.
type AuthMode int32

const (
	AuthMode_standard AuthMode = 0
	AuthMode_rsa1024  AuthMode = 1
	AuthMode_rsa2048  AuthMode = 2
)

// Enum value maps for AuthMode.
var (
	AuthMode_name = map[int32]string{
		0: "standard",
		1: "rsa1024",
		2: "rsa2048",
	}
	AuthMode_value = map[string]int32{
		"standard": 0,
		"rsa1024":  1,
		"rsa2048":  2,
	}
)

// ReportsMode is how daily/monthly reports are delivered.
type ReportsMode int32

const (
	ReportsMode_disabled   ReportsMode = 0
	ReportsMode_full       ReportsMode = 1
	ReportsMode_day_only   ReportsMode = 2
	ReportsMode_month_only ReportsMode = 3
)

// Enum value maps for ReportsMode.
var (
	ReportsMode_name = map[int32]string{
		0: "disabled",
		1: "full",
		2: "day_only",
		3: "month_only",
	}
	ReportsMode_value = map[string]int32{
		"disabled":   0,
		"full":       1,
		"day_only":   2,
		"month_only": 3,
	}
)

// ReportsFlags is a reports bitmask — combine with OR.
type ReportsFlags int32

const (
	ReportsFlags_none       ReportsFlags = 0
	ReportsFlags_email      ReportsFlags = 1
	ReportsFlags_support    ReportsFlags = 2
	ReportsFlags_statements ReportsFlags = 4
)

// Enum value maps for ReportsFlags.
var (
	ReportsFlags_name = map[int32]string{
		0: "none",
		1: "email",
		2: "support",
		4: "statements",
	}
	ReportsFlags_value = map[string]int32{
		"none":       0,
		"email":      1,
		"support":    2,
		"statements": 4,
	}
)

// NewsMode is how news is delivered to the group.
type NewsMode int32

const (
	NewsMode_disabled NewsMode = 0
	NewsMode_headers  NewsMode = 1
	NewsMode_full     NewsMode = 2
)

// Enum value maps for NewsMode.
var (
	NewsMode_name = map[int32]string{
		0: "disabled",
		1: "headers",
		2: "full",
	}
	NewsMode_value = map[string]int32{
		"disabled": 0,
		"headers":  1,
		"full":     2,
	}
)

// MailMode is how internal mail is delivered to the group.
type MailMode int32

const (
	MailMode_disabled MailMode = 0
	MailMode_full     MailMode = 1
)

// Enum value maps for MailMode.
var (
	MailMode_name = map[int32]string{
		0: "disabled",
		1: "full",
	}
	MailMode_value = map[string]int32{
		"disabled": 0,
		"full":     1,
	}
)

// GroupTradeFlags is the group-level trade bitmask — combine with OR.
// Distinct from SymbolTradeFlags on the symbol master.
type GroupTradeFlags int32

const (
	GroupTradeFlags_none                   GroupTradeFlags = 0x00000000
	GroupTradeFlags_swaps                  GroupTradeFlags = 0x00000001
	GroupTradeFlags_trailing               GroupTradeFlags = 0x00000002
	GroupTradeFlags_experts                GroupTradeFlags = 0x00000004
	GroupTradeFlags_expiration             GroupTradeFlags = 0x00000008
	GroupTradeFlags_signals_all            GroupTradeFlags = 0x00000010
	GroupTradeFlags_signals_own            GroupTradeFlags = 0x00000020
	GroupTradeFlags_so_compensation        GroupTradeFlags = 0x00000040
	GroupTradeFlags_so_fully_hedged        GroupTradeFlags = 0x00000080
	GroupTradeFlags_fifo_close             GroupTradeFlags = 0x00000100
	GroupTradeFlags_hedge_prohibit         GroupTradeFlags = 0x00000200
	GroupTradeFlags_deal_cost              GroupTradeFlags = 0x00000400
	GroupTradeFlags_so_compensation_credit GroupTradeFlags = 0x00000800
)

// Enum value maps for GroupTradeFlags.
var (
	GroupTradeFlags_name = map[int32]string{
		0x00000000: "none",
		0x00000001: "swaps",
		0x00000002: "trailing",
		0x00000004: "experts",
		0x00000008: "expiration",
		0x00000010: "signals_all",
		0x00000020: "signals_own",
		0x00000040: "so_compensation",
		0x00000080: "so_fully_hedged",
		0x00000100: "fifo_close",
		0x00000200: "hedge_prohibit",
		0x00000400: "deal_cost",
		0x00000800: "so_compensation_credit",
	}
	GroupTradeFlags_value = map[string]int32{
		"none":                   0x00000000,
		"swaps":                  0x00000001,
		"trailing":               0x00000002,
		"experts":                0x00000004,
		"expiration":             0x00000008,
		"signals_all":            0x00000010,
		"signals_own":            0x00000020,
		"so_compensation":        0x00000040,
		"so_fully_hedged":        0x00000080,
		"fifo_close":             0x00000100,
		"hedge_prohibit":         0x00000200,
		"deal_cost":              0x00000400,
		"so_compensation_credit": 0x00000800,
	}
)

// Has reports whether every bit in flag is set.
func (f GroupTradeFlags) Has(flag GroupTradeFlags) bool { return f&flag == flag }

// TransferMode is how balance transfers between accounts are allowed.
type TransferMode int32

const (
	TransferMode_disabled   TransferMode = 0
	TransferMode_by_name    TransferMode = 1
	TransferMode_group      TransferMode = 2
	TransferMode_name_group TransferMode = 3
)

// Enum value maps for TransferMode.
var (
	TransferMode_name = map[int32]string{
		0: "disabled",
		1: "by_name",
		2: "group",
		3: "name_group",
	}
	TransferMode_value = map[string]int32{
		"disabled":   0,
		"by_name":    1,
		"group":      2,
		"name_group": 3,
	}
)

// FreeMarginMode controls whether floating P/L counts toward free margin.
type FreeMarginMode int32

const (
	FreeMarginMode_not_use_pl FreeMarginMode = 0
	FreeMarginMode_use_pl     FreeMarginMode = 1
	FreeMarginMode_profit     FreeMarginMode = 2
	FreeMarginMode_loss       FreeMarginMode = 3
)

// Enum value maps for FreeMarginMode.
var (
	FreeMarginMode_name = map[int32]string{
		0: "not_use_pl",
		1: "use_pl",
		2: "profit",
		3: "loss",
	}
	FreeMarginMode_value = map[string]int32{
		"not_use_pl": 0,
		"use_pl":     1,
		"profit":     2,
		"loss":       3,
	}
)

// StopOutMode is the unit for Margin Call / Stop Out levels.
type StopOutMode int32

const (
	StopOutMode_percent StopOutMode = 0
	StopOutMode_money   StopOutMode = 1
)

// Enum value maps for StopOutMode.
var (
	StopOutMode_name = map[int32]string{
		0: "percent",
		1: "money",
	}
	StopOutMode_value = map[string]int32{
		"percent": 0,
		"money":   1,
	}
)

// MarginFreeProfitMode controls which floating P/L side counts as free.
type MarginFreeProfitMode int32

const (
	MarginFreeProfitMode_pl   MarginFreeProfitMode = 0
	MarginFreeProfitMode_loss MarginFreeProfitMode = 1
)

// Enum value maps for MarginFreeProfitMode.
var (
	MarginFreeProfitMode_name = map[int32]string{
		0: "pl",
		1: "loss",
	}
	MarginFreeProfitMode_value = map[string]int32{
		"pl":   0,
		"loss": 1,
	}
)

// MarginMode is the account margin calculation mode for the group.
type MarginMode int32

const (
	MarginMode_retail            MarginMode = 0
	MarginMode_exchange_discount MarginMode = 1
	MarginMode_retail_hedged     MarginMode = 2
)

// Enum value maps for MarginMode.
var (
	MarginMode_name = map[int32]string{
		0: "retail",
		1: "exchange_discount",
		2: "retail_hedged",
	}
	MarginMode_value = map[string]int32{
		"retail":            0,
		"exchange_discount": 1,
		"retail_hedged":     2,
	}
)

// GroupMarginFlags is the group-level margin bitmask — combine with OR.
// Distinct from SymbolMarginFlags on the symbol master.
type GroupMarginFlags int32

const (
	GroupMarginFlags_none      GroupMarginFlags = 0
	GroupMarginFlags_clear_acc GroupMarginFlags = 1
)

// Enum value maps for GroupMarginFlags.
var (
	GroupMarginFlags_name = map[int32]string{
		0: "none",
		1: "clear_acc",
	}
	GroupMarginFlags_value = map[string]int32{
		"none":      0,
		"clear_acc": 1,
	}
)

// HistoryLimit is how much deal/order history clients in the group can request.
type HistoryLimit int32

const (
	HistoryLimit_all      HistoryLimit = 0
	HistoryLimit_months_1 HistoryLimit = 1
	HistoryLimit_months_3 HistoryLimit = 2
	HistoryLimit_months_6 HistoryLimit = 3
	HistoryLimit_year_1   HistoryLimit = 4
	HistoryLimit_year_2   HistoryLimit = 5
	HistoryLimit_year_3   HistoryLimit = 6
)

// Enum value maps for HistoryLimit.
var (
	HistoryLimit_name = map[int32]string{
		0: "all",
		1: "months_1",
		2: "months_3",
		3: "months_6",
		4: "year_1",
		5: "year_2",
		6: "year_3",
	}
	HistoryLimit_value = map[string]int32{
		"all":      0,
		"months_1": 1,
		"months_3": 2,
		"months_6": 3,
		"year_1":   4,
		"year_2":   5,
		"year_3":   6,
	}
)

// Group is a config template row in hst.groups (path hierarchy).
type Group struct {
	GroupID  int64  `db:"group_id" json:"group_id"`
	Group    string `db:"group" json:"group"`
	Root     bool   `db:"root" json:"root"`
	ParentID *int64 `db:"parent_id" json:"parent_id,omitempty"`

	PermissionFlags PermissionsFlags `db:"permission_flags" json:"permission_flags"`
	AuthMode        AuthMode         `db:"auth_mode" json:"auth_mode"`
	AuthPasswordMin int32            `db:"auth_password_min" json:"auth_password_min"`

	Company             string `db:"company" json:"company"`
	CompanyPage         string `db:"company_page" json:"company_page"`
	CompanyEmail        string `db:"company_email" json:"company_email"`
	CompanySupportPage  string `db:"company_support_page" json:"company_support_page"`
	CompanySupportEmail string `db:"company_support_email" json:"company_support_email"`
	CompanyCatalog      string `db:"company_catalog" json:"company_catalog"`

	Currency       string `db:"currency" json:"currency"`
	CurrencyDigits int32  `db:"currency_digits" json:"currency_digits"`

	ReportsMode      ReportsMode  `db:"reports_mode" json:"reports_mode"`
	ReportsFlags     ReportsFlags `db:"reports_flags" json:"reports_flags"`
	ReportsEmail     string       `db:"reports_email" json:"reports_email"`
	ReportsSMTP      string       `db:"reports_smtp" json:"reports_smtp"`
	ReportsSMTPLogin string       `db:"reports_smtp_login" json:"reports_smtp_login"`

	NewsMode     NewsMode `db:"news_mode" json:"news_mode"`
	NewsCategory string   `db:"news_category" json:"news_category"`
	NewsLangs    []int32  `db:"news_langs" json:"news_langs"`
	MailMode     MailMode `db:"mail_mode" json:"mail_mode"`

	TradeFlags         GroupTradeFlags `db:"trade_flags" json:"trade_flags"`
	TradeInterestRate  float64         `db:"trade_interest_rate" json:"trade_interest_rate"`
	TradeVirtualCredit float64         `db:"trade_virtual_credit" json:"trade_virtual_credit"`
	TradeTransferMode  TransferMode    `db:"trade_transfer_mode" json:"trade_transfer_mode"`

	MarginFreeMode       FreeMarginMode       `db:"margin_free_mode" json:"margin_free_mode"`
	MarginSOMode         StopOutMode          `db:"margin_so_mode" json:"margin_so_mode"`
	MarginCall           float64              `db:"margin_call" json:"margin_call"`
	MarginStopOut        float64              `db:"margin_stop_out" json:"margin_stop_out"`
	MarginFreeProfitMode MarginFreeProfitMode `db:"margin_free_profit_mode" json:"margin_free_profit_mode"`
	MarginMode           MarginMode           `db:"margin_mode" json:"margin_mode"`
	MarginFlags          GroupMarginFlags     `db:"margin_flags" json:"margin_flags"`

	DemoLeverage *int32   `db:"demo_leverage" json:"demo_leverage"`
	DemoDeposit  *float64 `db:"demo_deposit" json:"demo_deposit"`

	LimitHistory         HistoryLimit `db:"limit_history" json:"limit_history"`
	LimitOrders          int32        `db:"limit_orders" json:"limit_orders"`
	LimitSymbols         int32        `db:"limit_symbols" json:"limit_symbols"`
	LimitPositions       int32        `db:"limit_positions" json:"limit_positions"`
	LimitPositionsVolume float64      `db:"limit_positions_volume" json:"limit_positions_volume"`

	UpdatedAt int64 `db:"updated_at" json:"updated_at"`
}

// GroupSymbol is a per-group symbol override in hst.groups_symbols.
// Nullable override fields mean inherit from the base symbol.
type GroupSymbol struct {
	SymbolID    int64  `db:"symbol_id" json:"symbol_id"`
	GroupID     int64  `db:"group_id" json:"group_id"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
	Path        string `db:"path" json:"path"`
	ConfigIndex int32  `db:"config_index" json:"config_index"`

	TradeMode         *TradeMode       `db:"trade_mode" json:"trade_mode,omitempty"`
	ExecMode          *ExecMode        `db:"exec_mode" json:"exec_mode,omitempty"`
	FillFlags         *FillingFlags    `db:"fill_flags" json:"fill_flags,omitempty"`
	ExpirFlags        *ExpirationFlags `db:"expir_flags" json:"expir_flags,omitempty"`
	SpreadDiff        *int32           `db:"spread_diff" json:"spread_diff,omitempty"`
	SpreadDiffBalance *int32           `db:"spread_diff_balance" json:"spread_diff_balance,omitempty"`
	StopsLevel        *int32           `db:"stops_level" json:"stops_level,omitempty"`
	FreezeLevel       *int32           `db:"freeze_level" json:"freeze_level,omitempty"`

	VolumeMin      *int64 `db:"volume_min" json:"volume_min,omitempty"`
	VolumeMinExt   *int64 `db:"volume_min_ext" json:"volume_min_ext,omitempty"`
	VolumeMax      *int64 `db:"volume_max" json:"volume_max,omitempty"`
	VolumeMaxExt   *int64 `db:"volume_max_ext" json:"volume_max_ext,omitempty"`
	VolumeStep     *int64 `db:"volume_step" json:"volume_step,omitempty"`
	VolumeStepExt  *int64 `db:"volume_step_ext" json:"volume_step_ext,omitempty"`
	VolumeLimit    *int64 `db:"volume_limit" json:"volume_limit,omitempty"`
	VolumeLimitExt *int64 `db:"volume_limit_ext" json:"volume_limit_ext,omitempty"`

	MarginFlags                    *SymbolMarginFlags `db:"margin_flags" json:"margin_flags,omitempty"`
	MarginInitial                  *float64           `db:"margin_initial" json:"margin_initial,omitempty"`
	MarginMaintenance              *float64           `db:"margin_maintenance" json:"margin_maintenance,omitempty"`
	MarginInitialBuy               *float64           `db:"margin_initial_buy" json:"margin_initial_buy,omitempty"`
	MarginInitialSell              *float64           `db:"margin_initial_sell" json:"margin_initial_sell,omitempty"`
	MarginInitialBuyLimit          *float64           `db:"margin_initial_buy_limit" json:"margin_initial_buy_limit,omitempty"`
	MarginInitialSellLimit         *float64           `db:"margin_initial_sell_limit" json:"margin_initial_sell_limit,omitempty"`
	MarginInitialBuyStop           *float64           `db:"margin_initial_buy_stop" json:"margin_initial_buy_stop,omitempty"`
	MarginInitialSellStop          *float64           `db:"margin_initial_sell_stop" json:"margin_initial_sell_stop,omitempty"`
	MarginInitialBuyStopLimit      *float64           `db:"margin_initial_buy_stop_limit" json:"margin_initial_buy_stop_limit,omitempty"`
	MarginInitialSellStopLimit     *float64           `db:"margin_initial_sell_stop_limit" json:"margin_initial_sell_stop_limit,omitempty"`
	MarginMaintenanceBuy           *float64           `db:"margin_maintenance_buy" json:"margin_maintenance_buy,omitempty"`
	MarginMaintenanceSell          *float64           `db:"margin_maintenance_sell" json:"margin_maintenance_sell,omitempty"`
	MarginMaintenanceBuyLimit      *float64           `db:"margin_maintenance_buy_limit" json:"margin_maintenance_buy_limit,omitempty"`
	MarginMaintenanceSellLimit     *float64           `db:"margin_maintenance_sell_limit" json:"margin_maintenance_sell_limit,omitempty"`
	MarginMaintenanceBuyStop       *float64           `db:"margin_maintenance_buy_stop" json:"margin_maintenance_buy_stop,omitempty"`
	MarginMaintenanceSellStop      *float64           `db:"margin_maintenance_sell_stop" json:"margin_maintenance_sell_stop,omitempty"`
	MarginMaintenanceBuyStopLimit  *float64           `db:"margin_maintenance_buy_stop_limit" json:"margin_maintenance_buy_stop_limit,omitempty"`
	MarginMaintenanceSellStopLimit *float64           `db:"margin_maintenance_sell_stop_limit" json:"margin_maintenance_sell_stop_limit,omitempty"`
	MarginCurrency                 *string            `db:"margin_currency" json:"margin_currency,omitempty"`
	MarginLiquidity                *float64           `db:"margin_liquidity" json:"margin_liquidity,omitempty"`
	MarginHedged                   *float64           `db:"margin_hedged" json:"margin_hedged,omitempty"`

	SwapMode          *SwapMode  `db:"swap_mode" json:"swap_mode,omitempty"`
	SwapLong          *float64   `db:"swap_long" json:"swap_long,omitempty"`
	SwapShort         *float64   `db:"swap_short" json:"swap_short,omitempty"`
	SwapYearDay       *int32     `db:"swap_year_day" json:"swap_year_day,omitempty"`
	SwapFlags         *SwapFlags `db:"swap_flags" json:"swap_flags,omitempty"`
	SwapRateSunday    *float64   `db:"swap_rate_sunday" json:"swap_rate_sunday,omitempty"`
	SwapRateMonday    *float64   `db:"swap_rate_monday" json:"swap_rate_monday,omitempty"`
	SwapRateTuesday   *float64   `db:"swap_rate_tuesday" json:"swap_rate_tuesday,omitempty"`
	SwapRateWednesday *float64   `db:"swap_rate_wednesday" json:"swap_rate_wednesday,omitempty"`
	SwapRateThursday  *float64   `db:"swap_rate_thursday" json:"swap_rate_thursday,omitempty"`
	SwapRateFriday    *float64   `db:"swap_rate_friday" json:"swap_rate_friday,omitempty"`
	SwapRateSaturday  *float64   `db:"swap_rate_saturday" json:"swap_rate_saturday,omitempty"`

	RETimeout *int32 `db:"re_timeout" json:"re_timeout,omitempty"`

	IECheckMode    *InstantMode `db:"ie_check_mode" json:"ie_check_mode,omitempty"`
	IETimeout      *int32       `db:"ie_timeout" json:"ie_timeout,omitempty"`
	IESlipProfit   *int32       `db:"ie_slip_profit" json:"ie_slip_profit,omitempty"`
	IESlipLosing   *int32       `db:"ie_slip_losing" json:"ie_slip_losing,omitempty"`
	IEVolumeMax    *int64       `db:"ie_volume_max" json:"ie_volume_max,omitempty"`
	IEVolumeMaxExt *int64       `db:"ie_volume_max_ext" json:"ie_volume_max_ext,omitempty"`
	IEFlags        *int32       `db:"ie_flags" json:"ie_flags,omitempty"`

	OrderFlags           *OrderFlags   `db:"order_flags" json:"order_flags,omitempty"`
	PermissionsFlags     *int32        `db:"permissions_flags" json:"permissions_flags,omitempty"`
	PermissionsBookDepth *int32        `db:"permissions_book_depth" json:"permissions_book_depth,omitempty"`
	REFlags              *RequestFlags `db:"re_flags" json:"re_flags,omitempty"`
}

// GroupStatusFromFlags maps permission_flags → active|inactive for admin UI.
// Zero flags = disabled; anything else (including default cert_confirm) = enabled.
func GroupStatusFromFlags(flags PermissionsFlags) string {
	if flags == PermissionsFlags_none {
		return "inactive"
	}
	return "active"
}

// PermissionFlagsFromStatus maps active|inactive → flags.
// Active uses the migration default (PERMISSION_CERT_CONFIRM = 1).
func PermissionFlagsFromStatus(status string) PermissionsFlags {
	switch status {
	case "inactive", "disabled", "Disabled":
		return PermissionsFlags_none
	default:
		return PermissionsFlags_cert_confirm
	}
}
