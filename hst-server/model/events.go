package model

// OrderEventType is the verb an order envelope carries.
type OrderEventType int32

const (
	OrderEventNew    OrderEventType = 1
	OrderEventUpdate OrderEventType = 2
	OrderEventCancel OrderEventType = 3
)

type OrderEvent struct {
	EventType OrderEventType `json:"event_type"`
	Data      *TradeRequest  `json:"data"`
}

// PositionEventType is the verb a position envelope carries.
type PositionEventType int32

const (
	PositionEventUpdate  PositionEventType = 1
	PositionEventClose   PositionEventType = 2
	PositionEventCloseBy PositionEventType = 3
)

type PositionEvent struct {
	EventType PositionEventType `json:"event_type"`
	Data      *TradeRequest     `json:"data"`
}

// DealingEventType is the verb a dealer's answer carries.
type DealingEventType int32

const (
	DealingEventOffer   DealingEventType = 1
	DealingEventConfirm DealingEventType = 2
	DealingEventRequote DealingEventType = 3
	DealingEventReject  DealingEventType = 4
	DealingEventCancel  DealingEventType = 5
	// DealingEventAccept is the client taking the price a dealer requoted them at.
	DealingEventAccept DealingEventType = 6
)

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

// TradeRequest is one request as the engine expects it.
type TradeRequest struct {
	RequestId    string  `json:"request_id"`
	Login        int64   `json:"login"`
	Symbol       string  `json:"symbol"`
	Type         int32   `json:"type"`
	Volume       int64   `json:"volume"`
	Price        float64 `json:"price"`
	PriceTrigger float64 `json:"price_trigger"`
	PriceSL      float64 `json:"price_sl"`
	PriceTP      float64 `json:"price_tp"`
	TypeFill     int32   `json:"type_fill"`
	TypeTime     int32   `json:"type_time"`
	Expiry       int64   `json:"expiry"`
	Deviation    int64   `json:"deviation"`
	OrderId      int64   `json:"order_id"`
	PositionId   int64   `json:"position_id"`
	PositionById int64   `json:"position_by_id"`
	Comment      string  `json:"comment"`
	ExpertId     int64   `json:"expert_id"`
	Reason       int32   `json:"reason"`
	Dealer       int64   `json:"dealer"`
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
