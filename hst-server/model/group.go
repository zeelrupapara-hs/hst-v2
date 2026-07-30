package model

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

	DemoLeverage int32   `db:"demo_leverage" json:"demo_leverage"`
	DemoDeposit  float64 `db:"demo_deposit" json:"demo_deposit"`

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

	TradeMode                      *TradeMode         `db:"trade_mode" json:"trade_mode,omitempty"`
	ExecMode                       *ExecMode          `db:"exec_mode" json:"exec_mode,omitempty"`
	FillFlags                      *FillingFlags      `db:"fill_flags" json:"fill_flags,omitempty"`
	ExpirFlags                     *ExpirationFlags   `db:"expir_flags" json:"expir_flags,omitempty"`
	SpreadDiff                     *int32             `db:"spread_diff" json:"spread_diff,omitempty"`
	SpreadDiffBalance              *int32             `db:"spread_diff_balance" json:"spread_diff_balance,omitempty"`
	StopsLevel                     *int32             `db:"stops_level" json:"stops_level,omitempty"`
	FreezeLevel                    *int32             `db:"freeze_level" json:"freeze_level,omitempty"`
	VolumeMin                      *int64             `db:"volume_min" json:"volume_min,omitempty"`
	VolumeMinExt                   *int64             `db:"volume_min_ext" json:"volume_min_ext,omitempty"`
	VolumeMax                      *int64             `db:"volume_max" json:"volume_max,omitempty"`
	VolumeMaxExt                   *int64             `db:"volume_max_ext" json:"volume_max_ext,omitempty"`
	VolumeStep                     *int64             `db:"volume_step" json:"volume_step,omitempty"`
	VolumeStepExt                  *int64             `db:"volume_step_ext" json:"volume_step_ext,omitempty"`
	VolumeLimit                    *int64             `db:"volume_limit" json:"volume_limit,omitempty"`
	VolumeLimitExt                 *int64             `db:"volume_limit_ext" json:"volume_limit_ext,omitempty"`
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
	SwapMode                       *SwapMode          `db:"swap_mode" json:"swap_mode,omitempty"`
	SwapLong                       *float64           `db:"swap_long" json:"swap_long,omitempty"`
	SwapShort                      *float64           `db:"swap_short" json:"swap_short,omitempty"`
	SwapYearDay                    *int32             `db:"swap_year_day" json:"swap_year_day,omitempty"`
	SwapFlags                      *SwapFlags         `db:"swap_flags" json:"swap_flags,omitempty"`
	SwapRateSunday                 *float64           `db:"swap_rate_sunday" json:"swap_rate_sunday,omitempty"`
	SwapRateMonday                 *float64           `db:"swap_rate_monday" json:"swap_rate_monday,omitempty"`
	SwapRateTuesday                *float64           `db:"swap_rate_tuesday" json:"swap_rate_tuesday,omitempty"`
	SwapRateWednesday              *float64           `db:"swap_rate_wednesday" json:"swap_rate_wednesday,omitempty"`
	SwapRateThursday               *float64           `db:"swap_rate_thursday" json:"swap_rate_thursday,omitempty"`
	SwapRateFriday                 *float64           `db:"swap_rate_friday" json:"swap_rate_friday,omitempty"`
	SwapRateSaturday               *float64           `db:"swap_rate_saturday" json:"swap_rate_saturday,omitempty"`
	RETimeout                      *int32             `db:"re_timeout" json:"re_timeout,omitempty"`
	IECheckMode                    *InstantMode       `db:"ie_check_mode" json:"ie_check_mode,omitempty"`
	IETimeout                      *int32             `db:"ie_timeout" json:"ie_timeout,omitempty"`
	IESlipProfit                   *int32             `db:"ie_slip_profit" json:"ie_slip_profit,omitempty"`
	IESlipLosing                   *int32             `db:"ie_slip_losing" json:"ie_slip_losing,omitempty"`
	IEVolumeMax                    *int64             `db:"ie_volume_max" json:"ie_volume_max,omitempty"`
	IEVolumeMaxExt                 *int64             `db:"ie_volume_max_ext" json:"ie_volume_max_ext,omitempty"`
	IEFlags                        *int32             `db:"ie_flags" json:"ie_flags,omitempty"`
	OrderFlags                     *OrderFlags        `db:"order_flags" json:"order_flags,omitempty"`
	PermissionsFlags               *int32             `db:"permissions_flags" json:"permissions_flags,omitempty"`
	PermissionsBookDepth           *int32             `db:"permissions_book_depth" json:"permissions_book_depth,omitempty"`
	REFlags                        *RequestFlags      `db:"re_flags" json:"re_flags,omitempty"`
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
