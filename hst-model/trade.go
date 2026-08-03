package model

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
