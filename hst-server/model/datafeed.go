package model

// FeederFlags is EnFeederFlags — data feed operation mode as a bitmask.
type FeederFlags int32

const (
	FeederFlags_quotes FeederFlags = 1
	FeederFlags_news   FeederFlags = 2
	FeederFlags_remote FeederFlags = 8
)

var (
	FeederFlags_name = map[int32]string{
		1: "quotes",
		2: "news",
		8: "remote",
	}
	FeederFlags_value = map[string]int32{
		"quotes": 1,
		"news":   2,
		"remote": 8,
	}
)

// FeederFieldFlags marks editable admin fields (API metadata, not stored).
type FeederFieldFlags int32

const (
	FeederFieldFlags_server FeederFieldFlags = 1
	FeederFieldFlags_login  FeederFieldFlags = 2
	FeederFieldFlags_pass   FeederFieldFlags = 4
	FeederFieldFlags_param  FeederFieldFlags = 8
)

// FeederParamType is the mt5_feeder_params.Type enumeration.
type FeederParamType int16

const (
	FeederParamType_string   FeederParamType = 0
	FeederParamType_int      FeederParamType = 1
	FeederParamType_float    FeederParamType = 2
	FeederParamType_time     FeederParamType = 3
	FeederParamType_date     FeederParamType = 4
	FeederParamType_datetime FeederParamType = 5
	FeederParamType_groups   FeederParamType = 6
	FeederParamType_symbols  FeederParamType = 7
	FeederParamType_bool     FeederParamType = 8
	FeederParamType_color    FeederParamType = 9
)

var (
	FeederParamType_name = map[int16]string{
		0: "string",
		1: "int",
		2: "float",
		3: "time",
		4: "date",
		5: "datetime",
		6: "groups",
		7: "symbols",
		8: "bool",
		9: "color",
	}
	FeederParamType_value = map[string]int16{
		"string":   0,
		"int":      1,
		"float":    2,
		"time":     3,
		"date":     4,
		"datetime": 5,
		"groups":   6,
		"symbols":  7,
		"bool":     8,
		"color":    9,
	}
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
	DatafeedID         int64                 `db:"datafeed_id" json:"datafeed_id"`
	Name               string                `db:"name" json:"name"`
	Module             string                `db:"module" json:"module"`
	Enable             DatafeedEnable        `db:"enable" json:"enable"`
	FeedIndex          int32                 `db:"feed_index" json:"feed_index"`
	AllowImportSymbols int16                 `db:"allow_import_symbols" json:"allow_import_symbols"`
	Mode               FeederFlags           `db:"mode" json:"mode"`
	GatewayServer      string                `db:"gateway_server" json:"gateway_server"`
	FeedServer         string                `db:"feed_server" json:"feed_server"`
	FeedLogin          int64                 `db:"feed_login" json:"feed_login"`
	FeedPassword       string                `db:"feed_password" json:"-"`
	GatewayLogin       int64                 `db:"gateway_login" json:"gateway_login"`
	GatewayPassword    string                `db:"gateway_password" json:"-"`
	Timeout            int32                 `db:"timeout" json:"timeout"`
	TimeoutReconnect   int32                 `db:"timeout_reconnect" json:"timeout_reconnect"`
	TimeoutSleep       int32                 `db:"timeout_sleep" json:"timeout_sleep"`
	AttemptsSleep      int32                 `db:"attempts_sleep" json:"attempts_sleep"`
	UpdatedAt          int64                 `db:"updated_at" json:"updated_at"`
	Company            string                `db:"company" json:"company"`
	Issuer             string                `db:"issuer" json:"issuer"`
	SysConnection      DatafeedSysConnection `db:"sys_connection" json:"sys_connection"`
	SysLastTime        int64                 `db:"sys_last_time" json:"sys_last_time"`
	TickStatsCount     int64                 `db:"tick_stats_count" json:"tick_stats_count"`
	TicksCount         int64                 `db:"ticks_count" json:"ticks_count"`
	BooksCount         int64                 `db:"books_count" json:"books_count"`
	NewsCount          int64                 `db:"news_count" json:"news_count"`
	BytesReceived      int64                 `db:"bytes_received" json:"bytes_received"`
	BytesSent          int64                 `db:"bytes_sent" json:"bytes_sent"`
	StateFlags         int32                 `db:"state_flags" json:"state_flags"`
}

// DatafeedParam is an additional data feed setting.
type DatafeedParam struct {
	ParamID    int64           `db:"param_id" json:"param_id"`
	DatafeedID int64           `db:"datafeed_id" json:"datafeed_id"`
	ParamKey   string          `db:"param_key" json:"param_key"`
	Type       FeederParamType `db:"type" json:"type"`
	Value      string          `db:"value" json:"value"`
	Priority   int32           `db:"priority" json:"priority"` // unique with datafeed_id, not globally
}

// DatafeedTranslate maps platform symbols to external source symbols.
type DatafeedTranslate struct {
	TranslateID int64  `db:"translate_id" json:"translate_id"`
	DatafeedID  int64  `db:"datafeed_id" json:"datafeed_id"`
	SymbolID    int64  `db:"symbol_id" json:"symbol_id"`
	Symbol      string `db:"symbol" json:"symbol"`
	Source      string `db:"source" json:"source"`
	BidMarkup   int32  `db:"bid_markup" json:"bid_markup"`
	AskMarkup   int32  `db:"ask_markup" json:"ask_markup"`
	Digits      int16  `db:"digits" json:"digits"`
}

// DatafeedSymbol is one MT5 Symbols tab row (explicit symbol or path mask rule).
type DatafeedSymbol struct {
	FeedSymbolID int64  `db:"feed_symbol_id" json:"feed_symbol_id"`
	DatafeedID   int64  `db:"datafeed_id" json:"datafeed_id"`
	SymbolID     *int64 `db:"symbol_id" json:"symbol_id,omitempty"`
	Path         string `db:"path" json:"path"`
	Exclude      int16  `db:"exclude" json:"exclude"`
	Symbol       string `db:"symbol" json:"symbol"`
}
