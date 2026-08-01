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

func BalanceActionName(action int32) string {
	if n, ok := DealAction_name[action]; ok {
		return n
	}
	return "balance"
}

// AffectsCredit reports whether the action moves credit rather than balance.
func AffectsCredit(action int32) bool {
	return action == int32(DealAction_credit) || action == int32(DealAction_bonus)
}

// IsBalanceAction reports whether an action is a money operation rather than a trade.
func IsBalanceAction(action int32) bool {
	switch DealAction(action) {
	case DealAction_balance, DealAction_credit, DealAction_charge, DealAction_correction,
		DealAction_bonus, DealAction_commission, DealAction_interest, DealAction_dividend,
		DealAction_dividend_franked, DealAction_tax, DealAction_agent,
		DealAction_so_compensation:
		return true
	}
	return false
}
