package model

type BalanceEventType int32

const (
	BalanceEventApply BalanceEventType = 1
)

type BalanceEvent struct {
	EventType BalanceEventType `json:"event_type"`
	Data      *BalanceRequest  `json:"data"`
}

// Amount is signed: a withdrawal, a charge and a negative correction are all negative.
type BalanceRequest struct {
	RequestId string  `json:"request_id"`
	Login     int64   `json:"login"`
	Action    int32   `json:"action"`
	Amount    float64 `json:"amount"`
	Comment   string  `json:"comment"`
	Dealer    int64   `json:"dealer"`
	ExpertId  int64   `json:"expert_id"`
	// Deposit refuses the operation when it would take the balance below zero.
	AllowNegative bool `json:"allow_negative"`
}

func BalanceActionName(action int32) string { return DealActionName(action) }

// AffectsCredit reports whether the action moves credit rather than balance.
func AffectsCredit(action int32) bool {
	return action == int32(DealCredit) || action == int32(DealBonus) ||
		action == int32(DealSOCompensationCredit)
}

// IsBalanceAction reports whether an action is a money operation rather than a trade.
func IsBalanceAction(action int32) bool {
	switch DealAction(action) {
	case DealBalance, DealCredit, DealCharge, DealCorrection, DealBonus,
		DealCommission, DealInterest, DealSOCompensation, DealSOCompensationCredit:
		return true
	}
	return false
}

func DealActionName(action int32) string {
	switch DealAction(action) {
	case DealBalance:
		return "balance"
	case DealCredit:
		return "credit"
	case DealCharge:
		return "charge"
	case DealCorrection:
		return "correction"
	case DealBonus:
		return "bonus"
	case DealCommission:
		return "commission"
	case DealInterest:
		return "interest"
	case DealSOCompensation:
		return "so_compensation"
	}
	return "balance"
}
