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

// DealAction is what the deal did to the account. Buy and sell trade; the rest move money.
type DealAction int32

const (
	DealAction_buy                    DealAction = 0
	DealAction_sell                   DealAction = 1
	DealAction_balance                DealAction = 2
	DealAction_credit                 DealAction = 3
	DealAction_charge                 DealAction = 4
	DealAction_correction             DealAction = 5
	DealAction_bonus                  DealAction = 6
	DealAction_commission             DealAction = 7
	DealAction_commission_daily       DealAction = 8
	DealAction_commission_monthly     DealAction = 9
	DealAction_agent_daily            DealAction = 10
	DealAction_agent_monthly          DealAction = 11
	DealAction_interest               DealAction = 12
	DealAction_buy_canceled           DealAction = 13
	DealAction_sell_canceled          DealAction = 14
	DealAction_dividend               DealAction = 15
	DealAction_dividend_franked       DealAction = 16
	DealAction_tax                    DealAction = 17
	DealAction_agent                  DealAction = 18
	DealAction_so_compensation        DealAction = 19
	DealAction_so_compensation_credit DealAction = 20
)

var (
	DealAction_name = map[int32]string{
		0:  "buy",
		1:  "sell",
		2:  "balance",
		3:  "credit",
		4:  "charge",
		5:  "correction",
		6:  "bonus",
		7:  "commission",
		8:  "commission_daily",
		9:  "commission_monthly",
		10: "agent_daily",
		11: "agent_monthly",
		12: "interest",
		13: "buy_canceled",
		14: "sell_canceled",
		15: "dividend",
		16: "dividend_franked",
		17: "tax",
		18: "agent",
		19: "so_compensation",
		20: "so_compensation_credit",
	}
	DealAction_value = valuesOf(DealAction_name)
)

// DealEntry says which way the deal moved the position.
type DealEntry int32

const (
	DealEntry_in     DealEntry = 0
	DealEntry_out    DealEntry = 1
	DealEntry_inout  DealEntry = 2
	DealEntry_out_by DealEntry = 3
)

var (
	DealEntry_name = map[int32]string{
		0: "in",
		1: "out",
		2: "inout",
		3: "out_by",
	}
	DealEntry_value = valuesOf(DealEntry_name)
)
