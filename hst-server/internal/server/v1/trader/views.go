package trader

// What a trader sees of its own orders, positions and deals.
//
// Volume is given in lots, the unit a trader works in. The engine's integer units are an
// internal detail and do not leave the server.

// ViewOrder is one working order.
type ViewOrder struct {
	OrderId        int64   `json:"order_id"`
	Symbol         string  `json:"symbol"`
	Type           int32   `json:"type"`
	State          int32   `json:"state"`
	Volume         float64 `json:"volume"`
	VolumeInitial  int64   `json:"-"`
	VolumeCurrent  int64   `json:"-"`
	PriceOrder     float64 `json:"price_order"`
	PriceCurrent   float64 `json:"price_current"`
	PriceSL        float64 `json:"price_sl"`
	PriceTP        float64 `json:"price_tp"`
	TimeSetup      int64   `json:"time_setup"`
	TimeExpiration int64   `json:"time_expiration"`
	Comment        string  `json:"comment"`
}

// ViewPosition is one open position.
type ViewPosition struct {
	PositionId   int64   `json:"position_id"`
	Symbol       string  `json:"symbol"`
	Action       int32   `json:"action"`
	Volume       float64 `json:"volume"`
	VolumeUnits  int64   `json:"-"`
	PriceOpen    float64 `json:"price_open"`
	PriceCurrent float64 `json:"price_current"`
	PriceSL      float64 `json:"price_sl"`
	PriceTP      float64 `json:"price_tp"`
	Profit       float64 `json:"profit"`
	Storage      float64 `json:"storage"`
	TimeCreate   int64   `json:"time_create"`
	Comment      string  `json:"comment"`
}

// ViewDeal is one line of the money history.
type ViewDeal struct {
	DealId      int64   `json:"deal_id"`
	OrderId     int64   `json:"order_id"`
	PositionId  int64   `json:"position_id"`
	Symbol      string  `json:"symbol"`
	Action      int32   `json:"action"`
	Entry       int32   `json:"entry"`
	Volume      float64 `json:"volume"`
	VolumeUnits int64   `json:"-"`
	Price       float64 `json:"price"`
	Profit      float64 `json:"profit"`
	Storage     float64 `json:"storage"`
	Commission  float64 `json:"commission"`
	Time        int64   `json:"time"`
	Comment     string  `json:"comment"`
}
