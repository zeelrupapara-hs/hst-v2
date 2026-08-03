package model

// CommissionMode is the commission type on a group commission setting.
type CommissionMode int32

const (
	CommissionMode_standard CommissionMode = 0
	CommissionMode_agent    CommissionMode = 1
)

var (
	CommissionMode_name = map[int32]string{
		0: "standard",
		1: "agent",
	}
	CommissionMode_value = map[string]int32{
		"standard": 0,
		"agent":    1,
	}
)

// CommissionRangeMode is how commission tiers are ranged.
type CommissionRangeMode int32

const (
	CommissionRangeMode_volume          CommissionRangeMode = 0
	CommissionRangeMode_turnover_money  CommissionRangeMode = 1
	CommissionRangeMode_turnover_volume CommissionRangeMode = 2
)

var (
	CommissionRangeMode_name = map[int32]string{
		0: "volume",
		1: "turnover_money",
		2: "turnover_volume",
	}
	CommissionRangeMode_value = map[string]int32{
		"volume":          0,
		"turnover_money":  1,
		"turnover_volume": 2,
	}
)

// CommissionChargeMode is when the commission is charged.
type CommissionChargeMode int32

const (
	CommissionChargeMode_daily   CommissionChargeMode = 0
	CommissionChargeMode_monthly CommissionChargeMode = 1
	CommissionChargeMode_instant CommissionChargeMode = 2
)

var (
	CommissionChargeMode_name = map[int32]string{
		0: "daily",
		1: "monthly",
		2: "instant",
	}
	CommissionChargeMode_value = map[string]int32{
		"daily":   0,
		"monthly": 1,
		"instant": 2,
	}
)

// CommissionEntryMode filters deals by entry direction.
type CommissionEntryMode int32

const (
	CommissionEntryMode_all CommissionEntryMode = 0
	CommissionEntryMode_in  CommissionEntryMode = 1
	CommissionEntryMode_out CommissionEntryMode = 2
)

var (
	CommissionEntryMode_name = map[int32]string{
		0: "all",
		1: "in",
		2: "out",
	}
	CommissionEntryMode_value = map[string]int32{
		"all": 0,
		"in":  1,
		"out": 2,
	}
)

// CommissionActionMode filters deals by buy/sell.
type CommissionActionMode int32

const (
	CommissionActionMode_all  CommissionActionMode = 0
	CommissionActionMode_buy  CommissionActionMode = 1
	CommissionActionMode_sell CommissionActionMode = 2
)

var (
	CommissionActionMode_name = map[int32]string{
		0: "all",
		1: "buy",
		2: "sell",
	}
	CommissionActionMode_value = map[string]int32{
		"all":  0,
		"buy":  1,
		"sell": 2,
	}
)

// CommissionProfitMode filters deals by profit/loss.
type CommissionProfitMode int32

const (
	CommissionProfitMode_all    CommissionProfitMode = 0
	CommissionProfitMode_profit CommissionProfitMode = 1
	CommissionProfitMode_loss   CommissionProfitMode = 2
)

var (
	CommissionProfitMode_name = map[int32]string{
		0: "all",
		1: "profit",
		2: "loss",
	}
	CommissionProfitMode_value = map[string]int32{
		"all":    0,
		"profit": 1,
		"loss":   2,
	}
)

// CommissionReasonFlags is a bitmask of deal reasons that attract commission.
type CommissionReasonFlags int32

const (
	CommissionReasonFlags_none     CommissionReasonFlags = 0x00000000
	CommissionReasonFlags_client   CommissionReasonFlags = 0x00000001
	CommissionReasonFlags_expert   CommissionReasonFlags = 0x00000002
	CommissionReasonFlags_dealer   CommissionReasonFlags = 0x00000004
	CommissionReasonFlags_external CommissionReasonFlags = 0x00000008
	CommissionReasonFlags_mobile   CommissionReasonFlags = 0x00000010
	CommissionReasonFlags_web      CommissionReasonFlags = 0x00000020
	CommissionReasonFlags_signal   CommissionReasonFlags = 0x00000040
	CommissionReasonFlags_gateway  CommissionReasonFlags = 0x00000080
	CommissionReasonFlags_ultency  CommissionReasonFlags = 0x00000100
)

var (
	CommissionReasonFlags_name = map[int32]string{
		0:   "none",
		1:   "client",
		2:   "expert",
		4:   "dealer",
		8:   "external",
		16:  "mobile",
		32:  "web",
		64:  "signal",
		128: "gateway",
		256: "ultency",
	}
	CommissionReasonFlags_value = map[string]int32{
		"none":     0,
		"client":   1,
		"expert":   2,
		"dealer":   4,
		"external": 8,
		"mobile":   16,
		"web":      32,
		"signal":   64,
		"gateway":  128,
		"ultency":  256,
	}
)

// CommissionTierMode is the unit for the tier value.
type CommissionTierMode int32

const (
	CommissionTierMode_deposit_currency CommissionTierMode = 0
	CommissionTierMode_base_currency    CommissionTierMode = 1
	CommissionTierMode_profit_currency  CommissionTierMode = 2
	CommissionTierMode_margin_currency  CommissionTierMode = 3
	CommissionTierMode_points           CommissionTierMode = 4
	CommissionTierMode_percent          CommissionTierMode = 5
	CommissionTierMode_specified        CommissionTierMode = 6
)

var (
	CommissionTierMode_name = map[int32]string{
		0: "deposit_currency",
		1: "base_currency",
		2: "profit_currency",
		3: "margin_currency",
		4: "points",
		5: "percent",
		6: "specified",
	}
	CommissionTierMode_value = map[string]int32{
		"deposit_currency": 0,
		"base_currency":    1,
		"profit_currency":  2,
		"margin_currency":  3,
		"points":           4,
		"percent":          5,
		"specified":        6,
	}
)

// CommissionTierType is per-trade vs per-volume charging.
type CommissionTierType int32

const (
	CommissionTierType_per_trade  CommissionTierType = 0
	CommissionTierType_per_volume CommissionTierType = 1
)

var (
	CommissionTierType_name = map[int32]string{
		0: "per_trade",
		1: "per_volume",
	}
	CommissionTierType_value = map[string]int32{
		"per_trade":  0,
		"per_volume": 1,
	}
)

// Commission is a group commission header.
type Commission struct {
	CommissionID     int64                 `db:"commission_id" json:"commission_id"`
	GroupID          int64                 `db:"group_id" json:"group_id"`
	Name             string                `db:"name" json:"name"`
	Description      string                `db:"description" json:"description"`
	Path             string                `db:"path" json:"path"`
	Mode             CommissionMode        `db:"mode" json:"mode"`
	ModeRange        CommissionRangeMode   `db:"mode_range" json:"mode_range"`
	ModeCharge       CommissionChargeMode  `db:"mode_charge" json:"mode_charge"`
	TurnoverCurrency string                `db:"turnover_currency" json:"turnover_currency"`
	ModeEntry        CommissionEntryMode   `db:"mode_entry" json:"mode_entry"`
	ModeAction       CommissionActionMode  `db:"mode_action" json:"mode_action"`
	ModeProfit       CommissionProfitMode  `db:"mode_profit" json:"mode_profit"`
	ModeReason       CommissionReasonFlags `db:"mode_reason" json:"mode_reason"`
	UpdatedAt        int64                 `db:"updated_at" json:"updated_at"`
}

// CommissionTier is one level under a commission.
type CommissionTier struct {
	TierID       int64              `db:"tier_id" json:"tier_id"`
	CommissionID int64              `db:"commission_id" json:"commission_id"`
	Mode         CommissionTierMode `db:"mode" json:"mode"`
	Type         CommissionTierType `db:"type" json:"type"`
	Value        float64            `db:"value" json:"value"`
	RangeFrom    float64            `db:"range_from" json:"range_from"`
	RangeTo      float64            `db:"range_to" json:"range_to"`
	Minimal      float64            `db:"minimal" json:"minimal"`
	Currency     string             `db:"currency" json:"currency"`
}
