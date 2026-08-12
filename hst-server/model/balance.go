package model

type BalanceEventType int32

const (
	BalanceEvent_apply BalanceEventType = 1
)

type BalanceEvent struct {
	EventType BalanceEventType `json:"event_type"`
	Data      *BalanceRequest  `json:"data"`
}

// Amount is signed: a withdrawal, a charge and a negative correction are all negative.
type BalanceRequest struct {
	RequestId string     `json:"request_id"`
	Login     int64      `json:"login"`
	Action    DealAction `json:"action"`
	Amount    float64    `json:"amount"`
	Comment   string     `json:"comment"`
	Dealer    int64      `json:"dealer"`
	ExpertId  int64      `json:"expert_id"`
	// Deposit refuses the operation when it would take the balance below zero.
	AllowNegative bool `json:"allow_negative"`
	// Fix writes Amount as the new value instead of adding it; no deal is recorded.
	Fix bool `json:"fix"`
}

func BalanceActionName(action DealAction) string { return DealActionName(action) }

// AffectsCredit reports whether the action moves credit rather than balance.
func AffectsCredit(action DealAction) bool {
	return action == DealAction_credit || action == DealAction_bonus ||
		action == DealAction_so_compensation_credit
}

// IsBalanceAction reports whether an action is a money operation rather than a trade.
func IsBalanceAction(action DealAction) bool {
	switch DealAction(action) {
	case DealAction_balance, DealAction_credit, DealAction_charge, DealAction_correction, DealAction_bonus,
		DealAction_commission, DealAction_interest, DealAction_so_compensation, DealAction_so_compensation_credit:
		return true
	}
	return false
}

func DealActionName(action DealAction) string {
	switch DealAction(action) {
	case DealAction_balance:
		return "balance"
	case DealAction_credit:
		return "credit"
	case DealAction_charge:
		return "charge"
	case DealAction_correction:
		return "correction"
	case DealAction_bonus:
		return "bonus"
	case DealAction_commission:
		return "commission"
	case DealAction_interest:
		return "interest"
	case DealAction_so_compensation:
		return "so_compensation"
	}
	return "balance"
}
