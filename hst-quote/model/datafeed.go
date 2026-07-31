package model

// FeederFlags is EnFeederFlags — data feed operation mode as a bitmask.
type FeederFlags int32

const (
	FeederFlags_quotes FeederFlags = 1
	FeederFlags_news   FeederFlags = 2
	FeederFlags_remote FeederFlags = 8
)

// DatafeedEnable is the enable flag on a data feed.
type DatafeedEnable int16

const (
	DatafeedEnable_disabled DatafeedEnable = 0
	DatafeedEnable_enabled  DatafeedEnable = 1
)

// DatafeedSysConnection is live connection state to the external source.
type DatafeedSysConnection int16

const (
	DatafeedSysConnection_down      DatafeedSysConnection = 0
	DatafeedSysConnection_connected DatafeedSysConnection = 1
)

// Datafeed is MT5 data feed configuration plus runtime session counters.
type Datafeed struct {
	DatafeedID       int64                 `db:"datafeed_id"`
	Name             string                `db:"name"`
	Module           string                `db:"module"`
	Enable           DatafeedEnable        `db:"enable"`
	Mode             FeederFlags           `db:"mode"`
	FeedServer       string                `db:"feed_server"`
	FeedLogin        int64                 `db:"feed_login"`
	FeedPassword     string                `db:"feed_password"`
	TimeoutReconnect int32                 `db:"timeout_reconnect"`
	SysConnection    DatafeedSysConnection `db:"sys_connection"`
	SysLastTime      int64                 `db:"sys_last_time"`
	TicksCount       int64                 `db:"ticks_count"`
	BytesReceived    int64                 `db:"bytes_received"`
}

// DatafeedParam is an additional data feed setting.
type DatafeedParam struct {
	ParamID    int64  `db:"param_id"`
	DatafeedID int64  `db:"datafeed_id"`
	ParamKey   string `db:"param_key"`
	Value      string `db:"value"`
}

// DatafeedTranslate maps platform symbol names to external source symbols.
type DatafeedTranslate struct {
	TranslateID int64  `db:"translate_id"`
	DatafeedID  int64  `db:"datafeed_id"`
	SymbolID    int64  `db:"symbol_id"`
	Symbol      string `db:"symbol"`
	Source      string `db:"source"`
	BidMarkup   int32  `db:"bid_markup"`
	AskMarkup   int32  `db:"ask_markup"`
	Digits      int16  `db:"digits"`
}

// QuoteFeed bundles a feed with params and symbol translations.
type QuoteFeed struct {
	Datafeed   Datafeed
	Params     map[string]string
	Translates []DatafeedTranslate
}

func (f QuoteFeed) Param(key string) string {
	if f.Params == nil {
		return ""
	}
	return f.Params[key]
}

// IsQuoteEnabled reports whether this feed should run in hst-quote.
func (f QuoteFeed) IsQuoteEnabled() bool {
	return f.Datafeed.Enable == DatafeedEnable_enabled &&
		(f.Datafeed.Mode&FeederFlags_quotes) != 0
}

// ExternalSymbol returns the LP symbol used for subscription.
func (t DatafeedTranslate) ExternalSymbol() string {
	if t.Source != "" {
		return t.Source
	}
	return t.Symbol
}
