package model

type QueryWhat int32

const (
	QueryAccount   QueryWhat = 1
	QueryPositions QueryWhat = 2
	QueryOrders    QueryWhat = 3
	QueryState     QueryWhat = 4
	QuerySymbols   QueryWhat = 5
)

type QueryRequest struct {
	Login int64     `json:"login"`
	What  QueryWhat `json:"what"`
}

type QueryResult struct {
	Login     int64        `json:"login"`
	Found     bool         `json:"found"`
	Account   *Account     `json:"account,omitempty"`
	Positions []Position   `json:"positions,omitempty"`
	Orders    []Order      `json:"orders,omitempty"`
	Symbols   []SymbolInfo `json:"symbols,omitempty"`
}

// SymbolInfo is one instrument as a terminal needs it: the group's resolved rules, not the
// instrument's own, plus the price it is quoted at right now.
//
// The order ticket cannot validate a volume without the limits, nor an SL without the stops
// level, and both are settled by the group override rather than the symbol row.
type SymbolInfo struct {
	Symbol      string `json:"symbol"`
	Path        string `json:"path"`
	Description string `json:"description"`

	Digits       int32   `json:"digits"`
	Point        float64 `json:"point"`
	ContractSize float64 `json:"contract_size"`
	TickValue    float64 `json:"tick_value"`
	TickSize     float64 `json:"tick_size"`
	CalcMode     int32   `json:"calc_mode"`
	TradeMode    int32   `json:"trade_mode"`
	ExecMode     int32   `json:"exec_mode"`
	FillFlags    int32   `json:"fill_flags"`
	ExpirFlags   int32   `json:"expir_flags"`
	OrderFlags   int32   `json:"order_flags"`

	// the volume limits are lots, already folded up from the extended units the engine counts in
	VolumeMin   float64 `json:"volume_min"`
	VolumeMax   float64 `json:"volume_max"`
	VolumeStep  float64 `json:"volume_step"`
	VolumeLimit float64 `json:"volume_limit"`

	// how far from the market an SL, TP or pending must sit, and how close it may no longer be moved
	StopsLevel  int32 `json:"stops_level"`
	FreezeLevel int32 `json:"freeze_level"`
	SpreadDiff  int32 `json:"spread_diff"`

	CurrencyBase   string `json:"currency_base"`
	CurrencyProfit string `json:"currency_profit"`
	CurrencyMargin string `json:"currency_margin"`

	MarginInitial     float64 `json:"margin_initial"`
	MarginMaintenance float64 `json:"margin_maintenance"`
	MarginHedged      float64 `json:"margin_hedged"`

	SwapMode  int32   `json:"swap_mode"`
	SwapLong  float64 `json:"swap_long"`
	SwapShort float64 `json:"swap_short"`

	// the live quote, zero when the instrument has not printed since the engine started
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	Last     float64 `json:"last"`
	Time     int64   `json:"time"`
	Gap      bool    `json:"gap"`
	HasQuote bool    `json:"has_quote"`
}
