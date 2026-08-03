package model

// What a trading account is told about its own trades.
//
// These mirror the API's ViewPosition, ViewOrder and ViewDeal field for field, because a client
// must not have to care whether a position arrived by REST or over the socket. In particular
// volume is in lots here, as it is everywhere a client can see; the internal unit count stays
// inside the engine.

// WirePosition is one open position as its owner sees it.
type WirePosition struct {
	PositionId   int64          `json:"position_id"`
	Login        int64          `json:"login"`
	Symbol       string         `json:"symbol"`
	Action       PositionAction `json:"action"`
	Reason       OrderReason    `json:"reason"`
	Volume       float64        `json:"volume"`
	PriceOpen    float64        `json:"price_open"`
	PriceCurrent float64        `json:"price_current"`
	PriceSL      float64        `json:"price_sl"`
	PriceTP      float64        `json:"price_tp"`
	Profit       float64        `json:"profit"`
	Storage      float64        `json:"storage"`
	TimeCreate   int64          `json:"time_create"`
	TimeUpdate   int64          `json:"time_update"`
	Comment      string         `json:"comment"`
}

func NewWirePosition(p *Position) *WirePosition {
	if p == nil {
		return nil
	}

	return &WirePosition{
		PositionId:   p.PositionId,
		Login:        p.Login,
		Symbol:       p.Symbol,
		Action:       p.Action,
		Reason:       p.Reason,
		Volume:       p.Lots(),
		PriceOpen:    p.PriceOpen,
		PriceCurrent: p.PriceCurrent,
		PriceSL:      p.PriceSL,
		PriceTP:      p.PriceTP,
		Profit:       p.Profit,
		Storage:      p.Storage,
		TimeCreate:   p.TimeCreate,
		TimeUpdate:   p.TimeUpdate,
		Comment:      p.Comment,
	}
}

// WireOrder is one order as its owner sees it.
type WireOrder struct {
	OrderId        int64       `json:"order_id"`
	Login          int64       `json:"login"`
	Symbol         string      `json:"symbol"`
	Type           OrderType   `json:"type"`
	State          OrderState  `json:"state"`
	Reason         OrderReason `json:"reason"`
	Volume         float64     `json:"volume"`
	PriceOrder     float64     `json:"price_order"`
	PriceTrigger   float64     `json:"price_trigger"`
	PriceCurrent   float64     `json:"price_current"`
	PriceSL        float64     `json:"price_sl"`
	PriceTP        float64     `json:"price_tp"`
	TimeSetup      int64       `json:"time_setup"`
	TimeExpiration int64       `json:"time_expiration"`
	Comment        string      `json:"comment"`
}

func NewWireOrder(o *Order) *WireOrder {
	if o == nil {
		return nil
	}

	return &WireOrder{
		OrderId:        o.OrderId,
		Login:          o.Login,
		Symbol:         o.Symbol,
		Type:           o.Type,
		State:          o.State,
		Reason:         o.Reason,
		Volume:         o.Lots(),
		PriceOrder:     o.PriceOrder,
		PriceTrigger:   o.PriceTrigger,
		PriceCurrent:   o.PriceCurrent,
		PriceSL:        o.PriceSL,
		PriceTP:        o.PriceTP,
		TimeSetup:      o.TimeSetup,
		TimeExpiration: o.TimeExpiration,
		Comment:        o.Comment,
	}
}

// WireDeal is one deal as its owner sees it.
type WireDeal struct {
	DealId     int64   `json:"deal_id"`
	Login      int64   `json:"login"`
	OrderId    int64   `json:"order_id"`
	PositionId int64   `json:"position_id"`
	Symbol     string  `json:"symbol"`
	Action     int32   `json:"action"`
	Entry      int32   `json:"entry"`
	Reason     int32   `json:"reason"`
	Volume     float64 `json:"volume"`
	Price      float64 `json:"price"`
	Profit     float64 `json:"profit"`
	Storage    float64 `json:"storage"`
	Commission float64 `json:"commission"`
	Fee        float64 `json:"fee"`
	Time       int64   `json:"time"`
	Comment    string  `json:"comment"`
}

func NewWireDeal(d *Deal) *WireDeal {
	if d == nil {
		return nil
	}

	return &WireDeal{
		DealId:     d.DealId,
		Login:      d.Login,
		OrderId:    d.OrderId,
		PositionId: d.PositionId,
		Symbol:     d.Symbol,
		Action:     int32(d.Action),
		Entry:      int32(d.Entry),
		Reason:     int32(d.Reason),
		Volume:     Lots(d.Volume),
		Price:      d.Price,
		Profit:     d.Profit,
		Storage:    d.Storage,
		Commission: d.Commission,
		Fee:        d.Fee,
		Time:       d.Time,
		Comment:    d.Comment,
	}
}
