package model

// EventType is what a websocket frame says happened. The value is the wire string.
type EventType string

const (
	EventOrderCreate   EventType = "order_create"
	EventOrderUpdate   EventType = "order_update"
	EventOrderCancel   EventType = "order_cancel"
	EventOrderRejected EventType = "order_rejected"
	EventOrderExpired  EventType = "order_expired"

	EventPositionCreate EventType = "position_create"
	EventPositionUpdate EventType = "position_update"
	EventPositionClose  EventType = "position_close"

	EventDealCreate EventType = "deal_create"

	EventAccountSummary EventType = "account_summary"
	EventMoneyChange    EventType = "money_change"
	EventMarginCall     EventType = "margin_call"
	EventStopOut        EventType = "stop_out"

	EventDealerRequest     EventType = "dealer_request"
	EventDealerRequestDone EventType = "dealer_request_done"

	EventJournalCreate EventType = "journal_create"
	EventMarketFeed    EventType = "market_feed"
)

// OrderEventType is the verb a command to the engine carries.
type OrderEventType int32

const (
	OrderEvent_new_order    OrderEventType = 1
	OrderEvent_update_order OrderEventType = 2
	OrderEvent_cancel_order OrderEventType = 3
)

var (
	OrderEvent_name = map[int32]string{
		1: "new_order",
		2: "update_order",
		3: "cancel_order",
	}
	OrderEvent_value = map[string]int32{
		"new_order":    1,
		"update_order": 2,
		"cancel_order": 3,
	}
)

// PositionEventType is the verb a position command carries.
type PositionEventType int32

const (
	PositionEvent_update   PositionEventType = 1
	PositionEvent_close    PositionEventType = 2
	PositionEvent_close_by PositionEventType = 3
)

var (
	PositionEvent_name = map[int32]string{
		1: "update",
		2: "close",
		3: "close_by",
	}
	PositionEvent_value = map[string]int32{
		"update":   1,
		"close":    2,
		"close_by": 3,
	}
)

// DealingEventType is the verb a dealer's answer carries.
type DealingEventType int32

const (
	DealingEvent_offer   DealingEventType = 1
	DealingEvent_confirm DealingEventType = 2
	DealingEvent_requote DealingEventType = 3
	DealingEvent_reject  DealingEventType = 4
	DealingEvent_cancel  DealingEventType = 5
	DealingEvent_accept  DealingEventType = 6
	DealingEvent_return  DealingEventType = 7
)

var (
	DealingEvent_name = map[int32]string{
		1: "offer",
		2: "confirm",
		3: "requote",
		4: "reject",
		5: "cancel",
		6: "accept",
		7: "return",
	}
	DealingEvent_value = map[string]int32{
		"offer":   1,
		"confirm": 2,
		"requote": 3,
		"reject":  4,
		"cancel":  5,
		"accept":  6,
		"return":  7,
	}
)
