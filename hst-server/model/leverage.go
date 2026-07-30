package model

// RangeMode is EnRangeMode, the level type for a floating leverage rule. It
// decides what quantity the rule's tiers are measured against.
type RangeMode int32

// The names follow the MT5 interface wording. RANGE_VALUE* are labelled
// "Notional value" there, which also keeps them from colliding with the
// RangeMode_value lookup map below.
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

// NeedsCurrency reports whether the mode measures notional value, in which case
// the rule must name the currency that value is converted to.
func (r RangeMode) NeedsCurrency() bool {
	return r == RangeMode_notional_value || r == RangeMode_notional_value_per_symbol
}

// Leverage is a floating leverage configuration. It does not replace the margin
// calculated from symbol settings, it multiplies it: the tiers of the matching
// rule carry rates that are applied as marginal brackets.
type Leverage struct {
	LeverageId int64  `db:"leverage_id" json:"leverage_id"`
	Name       string `db:"name" json:"name"`
	Timestamp  int64  `db:"timestamp" json:"timestamp"`
	Flags      int32  `db:"flags" json:"flags"`
}

func (Leverage) TableName() string { return "hst.leverages" }

// LeverageRule scopes a set of tiers to a symbol path mask. Rules are matched in
// config_index order and the first one matching an instrument wins, so the order
// is data rather than presentation.
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

// LeverageTier is one level of a rule. RangeTo 0 on the last tier means infinity.
// MarginRateMaintenance 0 means no margin is charged, not that the initial rate
// stands in for it, which is where this differs from symbol settings.
type LeverageTier struct {
	TierId                int64   `db:"tier_id" json:"tier_id"`
	RuleId                int64   `db:"rule_id" json:"rule_id"`
	RangeFrom             float64 `db:"range_from" json:"range_from"`
	RangeTo               float64 `db:"range_to" json:"range_to"`
	MarginRateInitial     float64 `db:"margin_rate_initial" json:"margin_rate_initial"`
	MarginRateMaintenance float64 `db:"margin_rate_maintenance" json:"margin_rate_maintenance"`
}

func (LeverageTier) TableName() string { return "hst.leverage_tiers" }
