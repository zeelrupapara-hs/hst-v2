package model

// RangeMode is EnRangeMode, the level type for a floating leverage rule.
type RangeMode int32

// The names follow the interface wording.
const (
	RangeMode_volume                    RangeMode = 0
	RangeMode_volume_per_symbol         RangeMode = 1
	RangeMode_notional_value            RangeMode = 2
	RangeMode_notional_value_per_symbol RangeMode = 3
)

// Enum value maps for RangeMode.
var (
	RangeMode_name = map[int32]string{
		0: "volume",
		1: "volume_per_symbol",
		2: "notional_value",
		3: "notional_value_per_symbol",
	}
	RangeMode_value = map[string]int32{
		"volume":                    0,
		"volume_per_symbol":         1,
		"notional_value":            2,
		"notional_value_per_symbol": 3,
	}
)

// NeedsCurrency reports whether the mode measures notional value.
func (r RangeMode) NeedsCurrency() bool {
	return r == RangeMode_notional_value || r == RangeMode_notional_value_per_symbol
}

// Leverage is a floating leverage configuration.
type Leverage struct {
	LeverageId int64  `db:"leverage_id" json:"leverage_id"`
	Name       string `db:"name" json:"name"`
	Timestamp  int64  `db:"timestamp" json:"timestamp"`
	Flags      int32  `db:"flags" json:"flags"`
}

func (Leverage) TableName() string { return "hst.leverages" }

// LeverageRule scopes a set of tiers to a symbol path mask.
type LeverageRule struct {
	RuleId                   int64     `db:"rule_id" json:"rule_id"`
	LeverageId               int64     `db:"leverage_id" json:"leverage_id"`
	Name                     string    `db:"name" json:"name"`
	Description              string    `db:"description" json:"description"`
	Path                     string    `db:"path" json:"path"`
	RangeMode                RangeMode `db:"range_mode" json:"range_mode"`
	RangeValueCurrency       string    `db:"range_value_currency" json:"range_value_currency"`
	RangeValueCurrencyDigits int32     `db:"range_value_currency_digits" json:"range_value_currency_digits"`
	ConfigIndex              int32     `db:"config_index" json:"config_index"`
}

func (LeverageRule) TableName() string { return "hst.leverage_rules" }

// LeverageTier is one level of a rule.
type LeverageTier struct {
	TierId                int64   `db:"tier_id" json:"tier_id"`
	RuleId                int64   `db:"rule_id" json:"rule_id"`
	RangeFrom             float64 `db:"range_from" json:"range_from"`
	RangeTo               float64 `db:"range_to" json:"range_to"`
	MarginRateInitial     float64 `db:"margin_rate_initial" json:"margin_rate_initial"`
	MarginRateMaintenance float64 `db:"margin_rate_maintenance" json:"margin_rate_maintenance"`
}

func (LeverageTier) TableName() string { return "hst.leverage_tiers" }
