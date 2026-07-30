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
