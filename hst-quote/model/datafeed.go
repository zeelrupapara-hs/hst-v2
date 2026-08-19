package model

import "math"

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
	// FeedIndex is the feed's position in the admin list; lower index wins symbol arbitration.
	FeedIndex int32       `db:"feed_index"`
	Mode      FeederFlags `db:"mode"`
	FeedServer       string                `db:"feed_server"`
	FeedLogin        string                `db:"feed_login"`
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

// QuoteFeed bundles a feed with params, symbol translations, and quote sessions.
type QuoteFeed struct {
	Datafeed   Datafeed
	Params     map[string]string
	Translates []DatafeedTranslate
	Sessions   []SymbolSession
	Settings   map[int64]SymbolSettings
}

// SymbolSettings is the per-symbol quote handling config from hst.symbols.
type SymbolSettings struct {
	SymbolID        int64
	Digits          int16
	Point           float64
	TickFlags       int32
	TickBookDepth   int32
	CalcMode        int16
	TickChartMode   int16
	SpliceType      int16
	FilterSoft      int32
	FilterSoftTicks int32
	FilterHard      int32
	FilterHardTicks int32
	FilterDiscard   int32
	FilterSpreadMin int32
	FilterSpreadMax int32
	FilterGap       int32
	FilterGapTicks  int32
	Spread          int32
	SpreadBalance   int32
}

// EnTickFlags — how the server treats incoming ticks for a symbol.
const (
	TickFlagRealtime   int32 = 1
	TickFlagCollectRaw int32 = 2
	TickFlagFeedStats  int32 = 4
)

// PointValue is the symbol point, falling back to digits when the column is unset.
func (s SymbolSettings) PointValue() float64 {
	if s.Point > 0 {
		return s.Point
	}
	if s.Digits > 0 {
		return math.Pow10(-int(s.Digits))
	}
	return 1e-5
}

// RealtimeAllowed treats an unconfigured tick_flags as "allow", never blackholing a feed.
func (s SymbolSettings) RealtimeAllowed() bool {
	return s.TickFlags == 0 || s.TickFlags&TickFlagRealtime != 0
}

func (s SymbolSettings) CollectRaw() bool { return s.TickFlags&TickFlagCollectRaw != 0 }

func (s SymbolSettings) HasDOM() bool { return s.TickBookDepth != 0 }

// FloatingSpread reports a feed-formed spread; market depth makes the fixed spread ignored.
func (s SymbolSettings) FloatingSpread() bool { return s.Spread == 0 || s.HasDOM() }

func (s SymbolSettings) IsExchange() bool { return s.CalcMode >= 32 && s.CalcMode <= 37 }

// FiltersEnabled reports whether the soft/hard/discard channel applies to this symbol.
func (s SymbolSettings) FiltersEnabled() bool {
	if s.FilterSoft == 0 && s.FilterHard == 0 && s.FilterDiscard == 0 {
		return false
	}
	if s.SpliceType != 0 {
		return false
	}
	if s.HasDOM() && (s.IsExchange() || s.TickChartMode == 1) {
		return false
	}
	return true
}

// SymbolSession is one quote session window for a symbol weekday.
type SymbolSession struct {
	SymbolID int64
	Day      int16
	Open     int32
	Close    int32
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
