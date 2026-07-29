package model

// Group is a config template row in hst.groups (path hierarchy).
type Group struct {
	GroupID   int64  `db:"group_id" json:"group_id"`
	UpdatedAt int64  `db:"updated_at" json:"updated_at"`
	Group     string `db:"group" json:"group"`
	Root      bool   `db:"root" json:"root"`
	ParentID  *int64 `db:"parent_id" json:"parent_id,omitempty"`

	PermissionFlags int32 `db:"permission_flags" json:"permission_flags"`
	AuthMode        int32 `db:"auth_mode" json:"auth_mode"`
	AuthPasswordMin int32 `db:"auth_password_min" json:"auth_password_min"`

	Company              string `db:"company" json:"company"`
	CompanyPage          string `db:"company_page" json:"company_page"`
	CompanyEmail         string `db:"company_email" json:"company_email"`
	CompanySupportPage   string `db:"company_support_page" json:"company_support_page"`
	CompanySupportEmail  string `db:"company_support_email" json:"company_support_email"`
	CompanyCatalog       string `db:"company_catalog" json:"company_catalog"`

	Currency       string `db:"currency" json:"currency"`
	CurrencyDigits int32  `db:"currency_digits" json:"currency_digits"`

	ReportsMode      int32  `db:"reports_mode" json:"reports_mode"`
	ReportsFlags     int32  `db:"reports_flags" json:"reports_flags"`
	ReportsEmail     string `db:"reports_email" json:"reports_email"`
	ReportsSMTP      string `db:"reports_smtp" json:"reports_smtp"`
	ReportsSMTPLogin string `db:"reports_smtp_login" json:"reports_smtp_login"`

	NewsMode     int32   `db:"news_mode" json:"news_mode"`
	NewsCategory string  `db:"news_category" json:"news_category"`
	NewsLangs    []int32 `db:"news_langs" json:"news_langs"`
	MailMode     int32   `db:"mail_mode" json:"mail_mode"`

	TradeFlags          int32   `db:"trade_flags" json:"trade_flags"`
	TradeInterestRate   float64 `db:"trade_interest_rate" json:"trade_interest_rate"`
	TradeVirtualCredit  float64 `db:"trade_virtual_credit" json:"trade_virtual_credit"`
	TradeTransferMode   int32   `db:"trade_transfer_mode" json:"trade_transfer_mode"`

	MarginFreeMode        int32   `db:"margin_free_mode" json:"margin_free_mode"`
	MarginSOMode          int32   `db:"margin_so_mode" json:"margin_so_mode"`
	MarginCall            float64 `db:"margin_call" json:"margin_call"`
	MarginStopOut         float64 `db:"margin_stop_out" json:"margin_stop_out"`
	MarginFreeProfitMode  int32   `db:"margin_free_profit_mode" json:"margin_free_profit_mode"`
	MarginMode            int32   `db:"margin_mode" json:"margin_mode"`
	MarginFlags           int32   `db:"margin_flags" json:"margin_flags"`

	DemoLeverage int32   `db:"demo_leverage" json:"demo_leverage"`
	DemoDeposit  float64 `db:"demo_deposit" json:"demo_deposit"`

	LimitHistory          int32   `db:"limit_history" json:"limit_history"`
	LimitOrders           int32   `db:"limit_orders" json:"limit_orders"`
	LimitSymbols          int32   `db:"limit_symbols" json:"limit_symbols"`
	LimitPositions        int32   `db:"limit_positions" json:"limit_positions"`
	LimitPositionsVolume  float64 `db:"limit_positions_volume" json:"limit_positions_volume"`
}

// GroupSymbol is a per-group symbol override in hst.groups_symbols.
// Nullable override fields mean inherit from the base symbol.
type GroupSymbol struct {
	SymbolID    int64  `db:"symbol_id" json:"symbol_id"`
	GroupID     int64  `db:"group_id" json:"group_id"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
	Path        string `db:"path" json:"path"`
	ConfigIndex int32  `db:"config_index" json:"config_index"`

	TradeMode *int32 `db:"trade_mode" json:"trade_mode,omitempty"`
	ExecMode  *int32 `db:"exec_mode" json:"exec_mode,omitempty"`
}

// GroupStatusFromFlags maps permission_flags → active|inactive for admin UI.
// Zero flags = disabled; anything else (including default 1) = enabled.
func GroupStatusFromFlags(flags int32) string {
	if flags == 0 {
		return "inactive"
	}
	return "active"
}

// PermissionFlagsFromStatus maps active|inactive → flags.
func PermissionFlagsFromStatus(status string) int32 {
	switch status {
	case "inactive", "disabled", "Disabled":
		return 0
	default:
		return 1
	}
}
