package model

type UsersRights int64

// bitmask, combine with OR. note trade_disabled is inverted: set means trading is off.
const (
	UsersRights_none              UsersRights = 0x0000
	UsersRights_enabled           UsersRights = 0x0001
	UsersRights_password          UsersRights = 0x0002
	UsersRights_trade_disabled    UsersRights = 0x0004
	UsersRights_investor          UsersRights = 0x0008
	UsersRights_confirmed         UsersRights = 0x0010
	UsersRights_trailing          UsersRights = 0x0020
	UsersRights_expert            UsersRights = 0x0040
	UsersRights_obsolete          UsersRights = 0x0080
	UsersRights_reports           UsersRights = 0x0100
	UsersRights_readonly          UsersRights = 0x0200
	UsersRights_reset_pass        UsersRights = 0x0400
	UsersRights_otp_enabled       UsersRights = 0x0800
	UsersRights_sponsored_hosting UsersRights = 0x2000
	UsersRights_api_enabled       UsersRights = 0x4000
	UsersRights_push_notification UsersRights = 0x8000
	UsersRights_technical         UsersRights = 0x10000
	UsersRights_exclude_reports   UsersRights = 0x20000
)

// 0x1000 is undocumented in the MT5 sources, treat as reserved.

// Enum value maps for UsersRights.
var (
	UsersRights_name = map[int64]string{
		0x0001:  "enabled",
		0x0002:  "password",
		0x0004:  "trade_disabled",
		0x0008:  "investor",
		0x0010:  "confirmed",
		0x0020:  "trailing",
		0x0040:  "expert",
		0x0080:  "obsolete",
		0x0100:  "reports",
		0x0200:  "readonly",
		0x0400:  "reset_pass",
		0x0800:  "otp_enabled",
		0x2000:  "sponsored_hosting",
		0x4000:  "api_enabled",
		0x8000:  "push_notification",
		0x10000: "technical",
		0x20000: "exclude_reports",
	}
	UsersRights_value = map[string]int64{
		"enabled":           0x0001,
		"password":          0x0002,
		"trade_disabled":    0x0004,
		"investor":          0x0008,
		"confirmed":         0x0010,
		"trailing":          0x0020,
		"expert":            0x0040,
		"obsolete":          0x0080,
		"reports":           0x0100,
		"readonly":          0x0200,
		"reset_pass":        0x0400,
		"otp_enabled":       0x0800,
		"sponsored_hosting": 0x2000,
		"api_enabled":       0x4000,
		"push_notification": 0x8000,
		"technical":         0x10000,
		"exclude_reports":   0x20000,
	}
)

// Has reports whether every bit in flag is set.
func (r UsersRights) Has(flag UsersRights) bool { return r&flag == flag }

// Set returns r with flag added.
func (r UsersRights) Set(flag UsersRights) UsersRights { return r | flag }

// Clear returns r with flag removed.
func (r UsersRights) Clear(flag UsersRights) UsersRights { return r &^ flag }

// CanConnect reports whether the account may log in at all.
func (r UsersRights) CanConnect() bool { return r.Has(UsersRights_enabled) }

// CanTrade folds the inverted trade_disabled flag so callers cannot get it backwards.
func (r UsersRights) CanTrade() bool {
	return r.Has(UsersRights_enabled) && !r.Has(UsersRights_trade_disabled)
}

// MustChangePassword reports whether the next login is limited to a password change.
func (r UsersRights) MustChangePassword() bool { return r.Has(UsersRights_reset_pass) }

type UsersPasswords int32

const (
	UsersPasswords_main     UsersPasswords = 0
	UsersPasswords_investor UsersPasswords = 1
	UsersPasswords_api      UsersPasswords = 2
)

// Enum value maps for UsersPasswords.
var (
	UsersPasswords_name = map[int32]string{
		0: "main",
		1: "investor",
		2: "api",
	}
	UsersPasswords_value = map[string]int32{
		"main":     0,
		"investor": 1,
		"api":      2,
	}
)

// User is a login, one trading account record. A client may own many.
type User struct {
	Login            int64       `db:"login" json:"login"`
	ClientId         int64       `db:"client_id" json:"client_id"`
	Group            string      `db:"group" json:"group"`
	Rights           UsersRights `db:"rights" json:"rights"`
	CertSerialNumber int64       `db:"cert_serial_number" json:"cert_serial_number"`

	Name       string `db:"name" json:"name"`
	FirstName  string `db:"first_name" json:"first_name"`
	LastName   string `db:"last_name" json:"last_name"`
	MiddleName string `db:"middle_name" json:"middle_name"`
	Company    string `db:"company" json:"company"`
	Account    string `db:"account" json:"account"`
	Country    string `db:"country" json:"country"`
	City       string `db:"city" json:"city"`
	State      string `db:"state" json:"state"`
	ZipCode    string `db:"zip_code" json:"zip_code"`
	Address    string `db:"address" json:"address"`
	Phone      string `db:"phone" json:"phone"`
	Email      string `db:"email" json:"email"`
	IdDocument string `db:"id_document" json:"id_document"`
	Status     string `db:"status" json:"status"`
	Comment    string `db:"comment" json:"comment"`
	Color      int32  `db:"color" json:"color"`
	Language   int32  `db:"language" json:"language"`
	Hsid       string `db:"hsid" json:"hsid"`

	// password hashes are never serialised out of the service
	PasswordMain     string `db:"password_main" json:"-"`
	PasswordInvestor string `db:"password_investor" json:"-"`
	PasswordApi      string `db:"password_api" json:"-"`
	PasswordPhone    string `db:"password_phone" json:"-"`
	FailedAttempts   int32  `db:"failed_attempts" json:"-"`
	LockedUntil      int64  `db:"locked_until" json:"-"`

	Leverage       int32 `db:"leverage" json:"leverage"`
	Agent          int64 `db:"agent" json:"agent"`
	LimitOrders    int32 `db:"limit_orders" json:"limit_orders"`
	LimitPositions int32 `db:"limit_positions" json:"limit_positions"`

	Balance           float64 `db:"balance" json:"balance"`
	Credit            float64 `db:"credit" json:"credit"`
	InterestRate      float64 `db:"interest_rate" json:"interest_rate"`
	CommissionDaily   float64 `db:"commission_daily" json:"commission_daily"`
	CommissionMonthly float64 `db:"commission_monthly" json:"commission_monthly"`
	BalancePrevDay    float64 `db:"balance_prev_day" json:"balance_prev_day"`
	BalancePrevMonth  float64 `db:"balance_prev_month" json:"balance_prev_month"`
	EquityPrevDay     float64 `db:"equity_prev_day" json:"equity_prev_day"`
	EquityPrevMonth   float64 `db:"equity_prev_month" json:"equity_prev_month"`

	TradeAccounts string `db:"trade_accounts" json:"trade_accounts"`
	LeadCampaign  string `db:"lead_campaign" json:"lead_campaign"`
	LeadSource    string `db:"lead_source" json:"lead_source"`
	ApiData       []byte `db:"api_data" json:"api_data"`

	Registration   int64  `db:"registration" json:"registration"`
	LastAccess     int64  `db:"last_access" json:"last_access"`
	LastPassChange int64  `db:"last_pass_change" json:"last_pass_change"`
	LastIp         string `db:"last_ip" json:"last_ip"`
	UpdatedAt      int64  `db:"updated_at" json:"updated_at"`
}

func (User) TableName() string { return "hst.users" }

// PasswordFor returns the hash held in the given slot.
func (u *User) PasswordFor(slot UsersPasswords) string {
	switch slot {
	case UsersPasswords_investor:
		return u.PasswordInvestor
	case UsersPasswords_api:
		return u.PasswordApi
	default:
		return u.PasswordMain
	}
}
