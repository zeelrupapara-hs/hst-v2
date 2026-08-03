package model

// Deal is one row of hst.deals.
type Deal struct {
	DealId          int64   `json:"deal_id"`
	ExternalId      string  `json:"external_id"`
	Login           int64   `json:"login"`
	Dealer          int64   `json:"dealer"`
	OrderId         int64   `json:"order_id"`
	Action          int32   `json:"action"`
	Entry           int32   `json:"entry"`
	Digits          int32   `json:"digits"`
	DigitsCurrency  int32   `json:"digits_currency"`
	ContractSize    float64 `json:"contract_size"`
	Time            int64   `json:"time"`
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price"`
	PriceSL         float64 `json:"price_sl"`
	PriceTP         float64 `json:"price_tp"`
	Volume          int64   `json:"volume"`
	VolumeExt       int64   `json:"volume_ext"`
	VolumeClosed    int64   `json:"volume_closed"`
	VolumeClosedExt int64   `json:"volume_closed_ext"`
	Profit          float64 `json:"profit"`
	Value           float64 `json:"value"`
	Storage         float64 `json:"storage"`
	Commission      float64 `json:"commission"`
	Fee             float64 `json:"fee"`
	RateProfit      float64 `json:"rate_profit"`
	RateMargin      float64 `json:"rate_margin"`
	ExpertId        int64   `json:"expert_id"`
	PositionId      int64   `json:"position_id"`
	Comment         string  `json:"comment"`
	ProfitRaw       float64 `json:"profit_raw"`
	PricePosition   float64 `json:"price_position"`
	TickValue       float64 `json:"tick_value"`
	TickSize        float64 `json:"tick_size"`
	Flags           int32   `json:"flags"`
	Reason          int32   `json:"reason"`
	Gateway         string  `json:"gateway"`
	PriceGateway    float64 `json:"price_gateway"`
	MarketBid       float64 `json:"market_bid"`
	MarketAsk       float64 `json:"market_ask"`
	MarketLast      float64 `json:"market_last"`
	ApiData         string  `json:"api_data"`
	DateCreated     int64   `json:"date_created"`
}
