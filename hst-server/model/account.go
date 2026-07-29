package model

// Account is the live money state of a login, one row per user.
// Floating must include storage and commission, or margin level reads high and stop-out fires late.
type Account struct {
	Login             int64   `db:"login" json:"login"`
	CurrencyDigits    int32   `db:"currency_digits" json:"currency_digits"`
	Balance           float64 `db:"balance" json:"balance"`
	Credit            float64 `db:"credit" json:"credit"`
	Margin            float64 `db:"margin" json:"margin"`
	MarginFree        float64 `db:"margin_free" json:"margin_free"`
	MarginLevel       float64 `db:"margin_level" json:"margin_level"`
	MarginLeverage    int32   `db:"margin_leverage" json:"margin_leverage"`
	MarginInitial     float64 `db:"margin_initial" json:"margin_initial"`
	MarginMaintenance float64 `db:"margin_maintenance" json:"margin_maintenance"`
	Profit            float64 `db:"profit" json:"profit"`
	Storage           float64 `db:"storage" json:"storage"`
	Floating          float64 `db:"floating" json:"floating"`
	Equity            float64 `db:"equity" json:"equity"`
	BlockedCommission float64 `db:"blocked_commission" json:"blocked_commission"`
	BlockedProfit     float64 `db:"blocked_profit" json:"blocked_profit"`
	Assets            float64 `db:"assets" json:"assets"`
	Liabilities       float64 `db:"liabilities" json:"liabilities"`
	UpdatedAt         int64   `db:"updated_at" json:"updated_at"`
}

func (Account) TableName() string { return "hst.accounts" }
