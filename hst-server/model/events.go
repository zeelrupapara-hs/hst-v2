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

	EventAlertTriggered      EventType = "alert_triggered"
	EventBadRequest          EventType = "bad_request"
	EventBalanceCreate       EventType = "balance_create"
	EventDealerCancel        EventType = "dealer_cancel"
	EventDealerConfirm       EventType = "dealer_confirm"
	EventDealerReject        EventType = "dealer_reject"
	EventDealerRequote       EventType = "dealer_requote"
	EventError               EventType = "error"
	EventForbidden           EventType = "forbidden"
	EventInternalServerError EventType = "internal_server_error"
	// EventMailInbox is what the trader terminal's socket handler listens for on a delivery.
	EventMailInbox             EventType = "email_inbox"
	EventNotFound              EventType = "not_found"
	EventOrderDealerCancel     EventType = "order_dealer_cancel"
	EventOrderDealerCreate     EventType = "order_dealer_create"
	EventOrderDealerUpdate     EventType = "order_dealer_update"
	EventPing                  EventType = "ping"
	EventPong                  EventType = "pong"
	EventPositionCloseBy       EventType = "position_close_by"
	EventPositionDealerClose   EventType = "position_dealer_close"
	EventPositionDealerCloseBy EventType = "position_dealer_close_by"
	EventPositionDealerUpdate  EventType = "position_dealer_update"
	EventSessionRevoked        EventType = "session.revoked"
	EventStartMarketFeed       EventType = "start_market_feed"
	EventStopMarketFeed        EventType = "stop_market_feed"
	EventUnauthorized          EventType = "unauthorized"
	EventWelcome               EventType = "welcome"
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
	// fix writes the deals-derived volume and open price back; no order, no deal
	PositionEvent_fix PositionEventType = 4
	// delete removes the position outright, the broker-level correction
	PositionEvent_delete PositionEventType = 5
)

var (
	PositionEvent_name = map[int32]string{
		1: "update",
		2: "close",
		3: "close_by",
		4: "fix",
		5: "delete",
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

// TradeRequest is one command as the engine expects it. Volume is in extended units.
type TradeRequest struct {
	RequestId    string       `json:"request_id"`
	Login        int64        `json:"login"`
	Symbol       string       `json:"symbol"`
	Type         OrderType    `json:"type"`
	Volume       int64        `json:"volume"`
	Price        float64      `json:"price"`
	PriceTrigger float64      `json:"price_trigger"`
	PriceSL      float64      `json:"price_sl"`
	PriceTP      float64      `json:"price_tp"`
	TypeFill     OrderFilling `json:"type_fill"`
	TypeTime     OrderTime    `json:"type_time"`
	ExpiryAt     int64        `json:"expiry_at"`
	Deviation    int64        `json:"deviation"`
	OrderId      int64        `json:"order_id"`
	PositionId   int64        `json:"position_id"`
	PositionById int64        `json:"position_by_id"`
	Comment      string       `json:"comment"`
	ExpertId     int64        `json:"expert_id"`
	Reason       OrderReason  `json:"reason"`
	Dealer       int64        `json:"dealer"`
}

// TradeResult is the engine's answer. Bid and Ask are set only on a requote.
type TradeResult struct {
	RequestId  string  `json:"request_id"`
	Login      int64   `json:"login"`
	RetCode    int32   `json:"retcode"`
	Message    string  `json:"message"`
	OrderId    int64   `json:"order_id,omitempty"`
	PositionId int64   `json:"position_id,omitempty"`
	DealId     int64   `json:"deal_id,omitempty"`
	Price      float64 `json:"price,omitempty"`
	Volume     int64   `json:"volume,omitempty"`
	Profit     float64 `json:"profit,omitempty"`
	Bid        float64 `json:"bid,omitempty"`
	Ask        float64 `json:"ask,omitempty"`
	Rule       string  `json:"rule,omitempty"`
}

// OrderEvent is a command envelope on system.orders.
type OrderEvent struct {
	EventType OrderEventType `json:"event_type"`
	Data      *TradeRequest  `json:"data"`
}

// PositionEvent is a command envelope on system.positions.
type PositionEvent struct {
	EventType PositionEventType `json:"event_type"`
	Data      *TradeRequest     `json:"data"`
}

// DealingEvent is a dealer's answer on system.dealing.
type DealingEvent struct {
	EventType DealingEventType `json:"event_type"`
	RequestId string           `json:"request_id"`
	Login     int64            `json:"login"`
	Dealer    int64            `json:"dealer"`
	Price     float64          `json:"price"`
	Reason    string           `json:"reason"`
	Request   *TradeRequest    `json:"request,omitempty"`
	At        int64            `json:"at"`
}
