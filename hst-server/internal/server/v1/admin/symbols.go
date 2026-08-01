package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "hstserver/internal/server/v1"
	"math"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CrtSymbolSession is one session window on create/replace.
type CrtSymbolSession struct {
	Type  int16 `json:"type" validate:"oneof=0 1"`
	Day   int16 `json:"day" validate:"gte=0,lte=6"`
	Open  int32 `json:"open" validate:"gte=0"`
	Close int32 `json:"close" validate:"gte=1,lte=1440"`
}

// CrtSymbol is the create payload, and point and multiply come from digits.
type CrtSymbol struct {
	Symbol         string             `json:"symbol" validate:"required,max=64"`
	Path           string             `json:"path" validate:"required,max=255"`
	Description    string             `json:"description" validate:"max=255"`
	ISIN           string             `json:"isin" validate:"max=32"`
	International  string             `json:"international" validate:"max=64"`
	Category       string             `json:"category" validate:"max=128"`
	Exchange       string             `json:"exchange" validate:"max=128"`
	Source         string             `json:"source" validate:"max=128"`
	CurrencyBase   string             `json:"currency_base" validate:"required,max=16"`
	CurrencyProfit string             `json:"currency_profit" validate:"required,max=16"`
	CurrencyMargin string             `json:"currency_margin" validate:"required,max=16"`
	Digits         int32              `json:"digits" validate:"gte=0,lte=12"`
	TradeMode      int16              `json:"trade_mode" validate:"gte=0,lte=4"`
	CalcMode       int16              `json:"calc_mode" validate:"gte=0"`
	ExecMode       int16              `json:"exec_mode" validate:"gte=0,lte=3"`
	ContractSize   float64            `json:"contract_size" validate:"gte=0"`
	Spread         int32              `json:"spread"`
	SpreadBalance  int32              `json:"spread_balance"`
	StopsLevel     int32              `json:"stops_level" validate:"gte=0"`
	FreezeLevel    int32              `json:"freeze_level" validate:"gte=0"`
	QuotesTimeout  int32              `json:"quotes_timeout" validate:"gte=0"`
	VolumeMin      int64              `json:"volume_min" validate:"gte=0"`
	VolumeMax      int64              `json:"volume_max" validate:"gte=0"`
	VolumeStep     int64              `json:"volume_step" validate:"gte=0"`
	TickSize       float64            `json:"tick_size" validate:"gte=0"`
	TickValue      float64            `json:"tick_value" validate:"gte=0"`
	SwapMode       int16              `json:"swap_mode" validate:"gte=0,lte=9"`
	SwapLong       float64            `json:"swap_long"`
	SwapShort      float64            `json:"swap_short"`
	FillFlags      int32              `json:"fill_flags" validate:"gte=0"`
	ExpirFlags     int32              `json:"expir_flags" validate:"gte=0"`
	OrderFlags     int32              `json:"order_flags" validate:"gte=0"`
	GTCMode        int16              `json:"gtc_mode" validate:"gte=0,lte=2"`
	Sessions       []CrtSymbolSession `json:"sessions" validate:"dive"`
}

// UptSymbol patches a symbol, and sending sessions replaces every row.
type UptSymbol struct {
	Symbol                         *string                  `json:"symbol"`
	Path                           *string                  `json:"path"`
	Isin                           *string                  `json:"isin"`
	Description                    *string                  `json:"description"`
	International                  *string                  `json:"international"`
	Category                       *string                  `json:"category"`
	Exchange                       *string                  `json:"exchange"`
	Cfi                            *string                  `json:"cfi"`
	Sector                         *model.SymbolSector      `json:"sector"`
	Industry                       *model.SymbolIndustry    `json:"industry"`
	Country                        *string                  `json:"country"`
	Basis                          *string                  `json:"basis"`
	Source                         *string                  `json:"source"`
	Page                           *string                  `json:"page"`
	CurrencyBase                   *string                  `json:"currency_base"`
	CurrencyBaseDigits             *int32                   `json:"currency_base_digits"`
	CurrencyProfit                 *string                  `json:"currency_profit"`
	CurrencyProfitDigits           *int32                   `json:"currency_profit_digits"`
	CurrencyMargin                 *string                  `json:"currency_margin"`
	CurrencyMarginDigits           *int32                   `json:"currency_margin_digits"`
	Color                          *int64                   `json:"color"`
	ColorBackground                *int64                   `json:"color_background"`
	Digits                         *int32                   `json:"digits"`
	TickFlags                      *model.TickFlags         `json:"tick_flags"`
	TickBookDepth                  *int32                   `json:"tick_book_depth"`
	FilterSoft                     *int32                   `json:"filter_soft"`
	FilterSoftTicks                *int32                   `json:"filter_soft_ticks"`
	FilterHard                     *int32                   `json:"filter_hard"`
	FilterHardTicks                *int32                   `json:"filter_hard_ticks"`
	FilterDiscard                  *int32                   `json:"filter_discard"`
	FilterSpreadMax                *int32                   `json:"filter_spread_max"`
	FilterSpreadMin                *int32                   `json:"filter_spread_min"`
	SubscriptionsDelay             *int32                   `json:"subscriptions_delay"`
	TradeMode                      *model.TradeMode         `json:"trade_mode"`
	CalcMode                       *model.CalcMode          `json:"calc_mode"`
	ExecMode                       *model.ExecMode          `json:"exec_mode"`
	GtcMode                        *model.GTCMode           `json:"gtc_mode"`
	FillFlags                      *model.FillingFlags      `json:"fill_flags"`
	ExpirFlags                     *model.ExpirationFlags   `json:"expir_flags"`
	Spread                         *int32                   `json:"spread"`
	SpreadBalance                  *int32                   `json:"spread_balance"`
	SpreadDiff                     *int32                   `json:"spread_diff"`
	SpreadDiffBalance              *int32                   `json:"spread_diff_balance"`
	TickValue                      *float64                 `json:"tick_value"`
	TickSize                       *float64                 `json:"tick_size"`
	ContractSize                   *float64                 `json:"contract_size"`
	StopsLevel                     *int32                   `json:"stops_level"`
	FreezeLevel                    *int32                   `json:"freeze_level"`
	QuotesTimeout                  *int32                   `json:"quotes_timeout"`
	VolumeMin                      *int64                   `json:"volume_min"`
	VolumeMinExt                   *int64                   `json:"volume_min_ext"`
	VolumeMax                      *int64                   `json:"volume_max"`
	VolumeMaxExt                   *int64                   `json:"volume_max_ext"`
	VolumeStep                     *int64                   `json:"volume_step"`
	VolumeStepExt                  *int64                   `json:"volume_step_ext"`
	VolumeLimit                    *int64                   `json:"volume_limit"`
	VolumeLimitExt                 *int64                   `json:"volume_limit_ext"`
	MarginFlags                    *model.SymbolMarginFlags `json:"margin_flags"`
	MarginInitial                  *float64                 `json:"margin_initial"`
	MarginMaintenance              *float64                 `json:"margin_maintenance"`
	MarginInitialBuy               *float64                 `json:"margin_initial_buy"`
	MarginInitialSell              *float64                 `json:"margin_initial_sell"`
	MarginInitialBuyLimit          *float64                 `json:"margin_initial_buy_limit"`
	MarginInitialSellLimit         *float64                 `json:"margin_initial_sell_limit"`
	MarginInitialBuyStop           *float64                 `json:"margin_initial_buy_stop"`
	MarginInitialSellStop          *float64                 `json:"margin_initial_sell_stop"`
	MarginInitialBuyStopLimit      *float64                 `json:"margin_initial_buy_stop_limit"`
	MarginInitialSellStopLimit     *float64                 `json:"margin_initial_sell_stop_limit"`
	MarginMaintenanceBuy           *float64                 `json:"margin_maintenance_buy"`
	MarginMaintenanceSell          *float64                 `json:"margin_maintenance_sell"`
	MarginMaintenanceBuyLimit      *float64                 `json:"margin_maintenance_buy_limit"`
	MarginMaintenanceSellLimit     *float64                 `json:"margin_maintenance_sell_limit"`
	MarginMaintenanceBuyStop       *float64                 `json:"margin_maintenance_buy_stop"`
	MarginMaintenanceSellStop      *float64                 `json:"margin_maintenance_sell_stop"`
	MarginMaintenanceBuyStopLimit  *float64                 `json:"margin_maintenance_buy_stop_limit"`
	MarginMaintenanceSellStopLimit *float64                 `json:"margin_maintenance_sell_stop_limit"`
	MarginHedged                   *float64                 `json:"margin_hedged"`
	SwapMode                       *model.SwapMode          `json:"swap_mode"`
	SwapLong                       *float64                 `json:"swap_long"`
	SwapShort                      *float64                 `json:"swap_short"`
	SwapYearDay                    *model.SwapDays          `json:"swap_year_day"`
	SwapFlags                      *model.SwapFlags         `json:"swap_flags"`
	SwapRateSunday                 *float64                 `json:"swap_rate_sunday"`
	SwapRateMonday                 *float64                 `json:"swap_rate_monday"`
	SwapRateTuesday                *float64                 `json:"swap_rate_tuesday"`
	SwapRateWednesday              *float64                 `json:"swap_rate_wednesday"`
	SwapRateThursday               *float64                 `json:"swap_rate_thursday"`
	SwapRateFriday                 *float64                 `json:"swap_rate_friday"`
	SwapRateSaturday               *float64                 `json:"swap_rate_saturday"`
	TimeStart                      *int64                   `json:"time_start"`
	TimeExpiration                 *int64                   `json:"time_expiration"`
	ReFlags                        *model.RequestFlags      `json:"re_flags"`
	ReTimeout                      *int32                   `json:"re_timeout"`
	IeCheckMode                    *model.InstantMode       `json:"ie_check_mode"`
	IeTimeout                      *int32                   `json:"ie_timeout"`
	IeSlipProfit                   *int32                   `json:"ie_slip_profit"`
	IeSlipLosing                   *int32                   `json:"ie_slip_losing"`
	IeVolumeMax                    *int64                   `json:"ie_volume_max"`
	IeVolumeMaxExt                 *int64                   `json:"ie_volume_max_ext"`
	PriceSettle                    *float64                 `json:"price_settle"`
	PriceLimitMax                  *float64                 `json:"price_limit_max"`
	PriceLimitMin                  *float64                 `json:"price_limit_min"`
	TradeFlags                     *model.SymbolTradeFlags  `json:"trade_flags"`
	OrderFlags                     *model.OrderFlags        `json:"order_flags"`
	MarginRateLiquidity            *float64                 `json:"margin_rate_liquidity"`
	MarginRateCurrency             *float64                 `json:"margin_rate_currency"`
	FaceValue                      *float64                 `json:"face_value"`
	AccruedInterest                *float64                 `json:"accrued_interest"`
	SpliceType                     *model.SpliceType        `json:"splice_type"`
	SpliceTimeType                 *model.SpliceTimeType    `json:"splice_time_type"`
	SpliceTimeDays                 *int32                   `json:"splice_time_days"`
	OptionMode                     *model.OptionMode        `json:"option_mode"`
	PriceStrike                    *float64                 `json:"price_strike"`
	FilterGap                      *int32                   `json:"filter_gap"`
	FilterGapTicks                 *int32                   `json:"filter_gap_ticks"`
	TickChartMode                  *model.ChartMode         `json:"tick_chart_mode"`
	Sessions                       *[]CrtSymbolSession      `json:"sessions"`
}

// ViewSymbolDetail is the full instrument plus sessions.
type ViewSymbolDetail struct {
	model.Symbol
	Sessions []model.SymbolSession `json:"sessions"`
}

var symbolsSortable = utils.NewSortable(
	"symbol_id", "symbol", "path", "digits", "trade_mode", "calc_mode",
	"exec_mode", "spread", "date_created", "date_modified")

const symbolListColumns = `symbol_id, symbol, path, description, digits, trade_mode,
	calc_mode, exec_mode, spread, contract_size, date_modified`

const symbolAllColumns = `
	symbol_id,
	symbol,
	path,
	isin,
	description,
	international,
	category,
	exchange,
	cfi,
	sector,
	industry,
	country,
	basis,
	source,
	page,
	currency_base,
	currency_base_digits,
	currency_profit,
	currency_profit_digits,
	currency_margin,
	currency_margin_digits,
	color,
	color_background,
	digits,
	point,
	multiply,
	tick_flags,
	tick_book_depth,
	filter_soft,
	filter_soft_ticks,
	filter_hard,
	filter_hard_ticks,
	filter_discard,
	filter_spread_max,
	filter_spread_min,
	subscriptions_delay,
	trade_mode,
	calc_mode,
	exec_mode,
	gtc_mode,
	fill_flags,
	expir_flags,
	spread,
	spread_balance,
	spread_diff,
	spread_diff_balance,
	tick_value,
	tick_size,
	contract_size,
	stops_level,
	freeze_level,
	quotes_timeout,
	volume_min,
	volume_min_ext,
	volume_max,
	volume_max_ext,
	volume_step,
	volume_step_ext,
	volume_limit,
	volume_limit_ext,
	margin_flags,
	margin_initial,
	margin_maintenance,
	margin_initial_buy,
	margin_initial_sell,
	margin_initial_buy_limit,
	margin_initial_sell_limit,
	margin_initial_buy_stop,
	margin_initial_sell_stop,
	margin_initial_buy_stop_limit,
	margin_initial_sell_stop_limit,
	margin_maintenance_buy,
	margin_maintenance_sell,
	margin_maintenance_buy_limit,
	margin_maintenance_sell_limit,
	margin_maintenance_buy_stop,
	margin_maintenance_sell_stop,
	margin_maintenance_buy_stop_limit,
	margin_maintenance_sell_stop_limit,
	margin_hedged,
	swap_mode,
	swap_long,
	swap_short,
	swap_year_day,
	swap_flags,
	swap_rate_sunday,
	swap_rate_monday,
	swap_rate_tuesday,
	swap_rate_wednesday,
	swap_rate_thursday,
	swap_rate_friday,
	swap_rate_saturday,
	time_start,
	time_expiration,
	re_flags,
	re_timeout,
	ie_check_mode,
	ie_timeout,
	ie_slip_profit,
	ie_slip_losing,
	ie_volume_max,
	ie_volume_max_ext,
	price_settle,
	price_limit_max,
	price_limit_min,
	trade_flags,
	order_flags,
	margin_rate_liquidity,
	margin_rate_currency,
	face_value,
	accrued_interest,
	splice_type,
	splice_time_type,
	splice_time_days,
	option_mode,
	price_strike,
	filter_gap,
	filter_gap_ticks,
	tick_chart_mode,
	date_created,
	date_modified`

func pointMultiply(digits int32) (float64, float64) {
	return math.Pow10(int(-digits)), math.Pow10(int(digits))
}

// symbolPath puts the symbol at the end of its folder, the way MT5 stores it.
func symbolPath(folder, symbol string) string {
	folder = strings.Trim(folder, `\`)
	if folder == "" || folder == symbol {
		return symbol
	}
	// the caller may already have sent the full path
	if i := strings.LastIndex(folder, `\`); i >= 0 && folder[i+1:] == symbol {
		return folder
	}
	return folder + `\` + symbol
}

var weekdayNames = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// sessionWindow is one session row, in the shape the log prints it.
type sessionWindow struct {
	Type  int16
	Day   int16
	Open  int32
	Close int32
}

// String prints the window the way a person reads it, clock time and all.
func (w sessionWindow) String() string {
	kind := "quote"
	if w.Type == int16(model.SymbolSessionType_trade) {
		kind = "trade"
	}

	day := "?"
	if w.Day >= 0 && int(w.Day) < len(weekdayNames) {
		day = weekdayNames[w.Day]
	}

	return fmt.Sprintf("%s %s %02d:%02d-%02d:%02d", kind, day,
		w.Open/60, w.Open%60, w.Close/60, w.Close%60)
}

// sortSessions puts the windows in a fixed order, so a reorder is not a change.
func sortSessions(windows []sessionWindow) {
	sort.Slice(windows, func(i, j int) bool {
		a, b := windows[i], windows[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Day != b.Day {
			return a.Day < b.Day
		}
		return a.Open < b.Open
	})
}

// printSessions joins the windows into one bracketed list.
func printSessions(windows []sessionWindow) string {
	out := make([]string, len(windows))
	for i, w := range windows {
		out[i] = w.String()
	}
	return "[" + strings.Join(out, ", ") + "]"
}

// jsonString reads a json string value, and returns "" for anything else.
func jsonString(raw json.RawMessage) string {
	var out string
	_ = json.Unmarshal(raw, &out)
	return out
}

// sameJSON compares two json values, treating 100 and 100.00000000 as equal.
func sameJSON(was, next json.RawMessage) bool {
	if bytes.Equal(was, next) {
		return true
	}
	oldNum, oldErr := strconv.ParseFloat(string(was), 64)
	newNum, newErr := strconv.ParseFloat(string(next), 64)
	return oldErr == nil && newErr == nil && oldNum == newNum
}

// changedSymbolFields lists what the request really changes, as key=value.
// A field the caller sent with the value it already had is left out.
func changedSymbolFields(before map[string]json.RawMessage, body *UptSymbol, sessions []sessionWindow) string {
	changes := make([]string, 0, 8)
	fields := reflect.ValueOf(*body)
	types := fields.Type()

	for i := range fields.NumField() {
		field := fields.Field(i)
		if field.Kind() != reflect.Pointer || field.IsNil() {
			continue
		}

		name, _, _ := strings.Cut(types.Field(i).Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}

		// sessions replace every row, so the count says more than the whole list
		// sessions replace every row, so the whole new list is the value
		if name == "sessions" {
			sent, ok := field.Interface().(*[]CrtSymbolSession)
			if !ok {
				continue
			}

			next := make([]sessionWindow, 0, len(*sent))
			for _, w := range *sent {
				next = append(next, sessionWindow(w))
			}
			sortSessions(next)

			if !slices.Equal(sessions, next) {
				changes = append(changes, "sessions="+printSessions(next))
			}
			continue
		}

		next, err := json.Marshal(field.Elem().Interface())
		if err != nil || sameJSON(before[name], next) {
			continue
		}

		changes = append(changes, name+"="+string(next))
	}

	return strings.Join(changes, ", ")
}

// symbolFolder drops the symbol from the end of a stored path.
func symbolFolder(path string) string {
	if i := strings.LastIndex(path, `\`); i >= 0 {
		return path[:i]
	}
	return ""
}

func validateSessions(sessions []CrtSymbolSession) error {
	type key struct {
		t, d int16
	}
	by := map[key][]CrtSymbolSession{}
	for _, s := range sessions {
		if s.Open >= s.Close {
			return fmt.Errorf("session open must be < close")
		}
		if s.Close > 1440 || s.Open < 0 {
			return fmt.Errorf("session open/close out of range")
		}
		k := key{s.Type, s.Day}
		by[k] = append(by[k], s)
	}
	for _, list := range by {
		sort.Slice(list, func(i, j int) bool { return list[i].Open < list[j].Open })
		for i := 1; i < len(list); i++ {
			if list[i].Open < list[i-1].Close {
				return fmt.Errorf("overlapping sessions for type=%d day=%d", list[i].Type, list[i].Day)
			}
		}
	}
	return nil
}

// CreateSymbol registers a server-wide instrument.
//
//	@Id			CreateSymbol
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtSymbol	true	"the symbol to create"
//	@Success	201		{object}	Response{data=v1.ViewSymbol}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols [post]
func (s *Server) CreateSymbol(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var body CrtSymbol
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.VolumeMax > 0 && body.VolumeMin > 0 && body.VolumeMax < body.VolumeMin {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("volume_max must be >= volume_min"))
	}
	if body.VolumeStep == 0 && (body.VolumeMin > 0 || body.VolumeMax > 0) {
		body.VolumeStep = 10000
	}
	if err := validateSessions(body.Sessions); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// zero means "use schema-like defaults" for create (patch can set disabled=0 later)
	tradeMode := body.TradeMode
	if tradeMode == 0 {
		tradeMode = 4
	}
	execMode := body.ExecMode
	if execMode == 0 {
		execMode = 2
	}
	contractSize := body.ContractSize
	if contractSize == 0 {
		contractSize = 100000
	}
	volumeMin := body.VolumeMin
	if volumeMin == 0 {
		volumeMin = 10000
	}
	volumeMax := body.VolumeMax
	if volumeMax == 0 {
		volumeMax = 100000000
	}
	volumeStep := body.VolumeStep
	if volumeStep == 0 {
		volumeStep = 10000
	}
	if volumeMax < volumeMin {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("volume_max must be >= volume_min"))
	}

	point, multiply := pointMultiply(body.Digits)
	path := symbolPath(body.Path, body.Symbol)
	if len(path) > 255 {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("path and symbol are longer than 255 together"))
	}
	now := time.Now().UnixNano()

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	view := &v1.ViewSymbol{}
	err = tx.QueryRow(ctx,
		`INSERT INTO hst.symbols
		   (symbol, path, description, isin, international, category, exchange, source,
		    currency_base, currency_profit, currency_margin, digits, point, multiply,
		    trade_mode, calc_mode, exec_mode, contract_size, spread, spread_balance,
		    stops_level, freeze_level, quotes_timeout, volume_min, volume_max, volume_step,
		    tick_size, tick_value, swap_mode, swap_long, swap_short,
		    fill_flags, expir_flags, order_flags, gtc_mode,
		    date_created, date_modified)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
		         $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$36)
		 RETURNING `+symbolListColumns,
		body.Symbol, path, body.Description, body.ISIN, body.International,
		body.Category, body.Exchange, body.Source,
		body.CurrencyBase, body.CurrencyProfit, body.CurrencyMargin,
		body.Digits, point, multiply,
		tradeMode, body.CalcMode, execMode, contractSize, body.Spread, body.SpreadBalance,
		body.StopsLevel, body.FreezeLevel, body.QuotesTimeout,
		volumeMin, volumeMax, volumeStep,
		body.TickSize, body.TickValue, body.SwapMode, body.SwapLong, body.SwapShort,
		body.FillFlags, body.ExpirFlags, body.OrderFlags, body.GTCMode,
		now).
		Scan(&view.SymbolId, &view.Symbol, &view.Path, &view.Description, &view.Digits,
			&view.TradeMode, &view.CalcMode, &view.ExecMode, &view.Spread, &view.ContractSize,
			&view.DateModified)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if err := insertSessions(ctx, tx, view.SymbolId, body.Sessions); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "symbol created",
		"actor", snap.Login, "symbol_id", view.SymbolId,
		"symbol", view.Symbol, "path", view.Path)

	s.NotifyWS(model.SubjectSymbol, model.EventSymbolCreated, view)
	s.NotifySystem(model.SubjectSystemSymbolCreated, view)
	s.JournalEntry(c, logger.CodeOK, journal.SymbolCreatedMsg(snap.Login, view.Symbol), view)

	return s.App.HttpResponseCreated(c, view)
}

type sessionInserter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertSessions(ctx context.Context, db sessionInserter, symbolID int64, sessions []CrtSymbolSession) error {
	for _, sess := range sessions {
		if _, err := db.Exec(ctx,
			`INSERT INTO hst.symbols_sessions (symbol_id, type, day, open, close)
			 VALUES ($1,$2,$3,$4,$5)`,
			symbolID, sess.Type, sess.Day, sess.Open, sess.Close); err != nil {
			return err
		}
	}
	return nil
}

// ListSymbols returns a page of symbols.
//
//	@Id			ListSymbols
//	@Tags		Symbols
//	@Produce	json
//	@Param		page	query		int		false	"page number, from 1"
//	@Param		limit	query		int		false	"rows per page, max 500"
//	@Param		search	query		string	false	"matches symbol, path or description"
//	@Param		sort_by	query		string	false	"symbol_id, symbol, path, digits, trade_mode, calc_mode, exec_mode, spread, date_created, date_modified"	Enums(symbol_id, symbol, path, digits, trade_mode, calc_mode, exec_mode, spread, date_created, date_modified)
//	@Param		order	query		string	false	"asc or desc"																								Enums(asc, desc)
//	@Success	200		{object}	Response{data=[]v1.ViewSymbol}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols [get]
func (s *Server) ListSymbols(c *fiber.Ctx) error {
	q, err := utils.QueryFilter(c, symbolsSortable, "symbol")
	if err != nil {
		return s.App.HttpResponseBadQueryParams(c, err)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+symbolListColumns+`
		   FROM hst.symbols
		  WHERE ($1 = '' OR symbol ILIKE '%'||$1||'%' OR path ILIKE '%'||$1||'%'
		         OR description ILIKE '%'||$1||'%')
		  ORDER BY `+q.SortBy+`
		  LIMIT $2 OFFSET $3`, q.Search, q.Limit, q.Offset)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []v1.ViewSymbol{}
	for rows.Next() {
		var v v1.ViewSymbol
		if err := rows.Scan(&v.SymbolId, &v.Symbol, &v.Path, &v.Description, &v.Digits,
			&v.TradeMode, &v.CalcMode, &v.ExecMode, &v.Spread, &v.ContractSize,
			&v.DateModified); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}
	return s.App.HttpResponseOK(c, out)
}

// GetSymbol returns one symbol with sessions.
//
//	@Id			GetSymbol
//	@Tags		Symbols
//	@Produce	json
//	@Param		id	path		int	true	"symbol id"
//	@Success	200	{object}	Response{data=ViewSymbolDetail}
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/{id} [get]
func (s *Server) GetSymbol(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	detail, err := s.selectSymbolDetail(c.UserContext(), int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, detail)
}

func (s *Server) selectSymbolDetail(ctx context.Context, id int64) (*ViewSymbolDetail, error) {
	sym := model.Symbol{}
	err := s.DB.DB.QueryRow(ctx,
		`SELECT `+symbolAllColumns+` FROM hst.symbols WHERE symbol_id = $1`, id).
		Scan(
			&sym.SymbolId,
			&sym.Symbol,
			&sym.Path,
			&sym.Isin,
			&sym.Description,
			&sym.International,
			&sym.Category,
			&sym.Exchange,
			&sym.Cfi,
			&sym.Sector,
			&sym.Industry,
			&sym.Country,
			&sym.Basis,
			&sym.Source,
			&sym.Page,
			&sym.CurrencyBase,
			&sym.CurrencyBaseDigits,
			&sym.CurrencyProfit,
			&sym.CurrencyProfitDigits,
			&sym.CurrencyMargin,
			&sym.CurrencyMarginDigits,
			&sym.Color,
			&sym.ColorBackground,
			&sym.Digits,
			&sym.Point,
			&sym.Multiply,
			&sym.TickFlags,
			&sym.TickBookDepth,
			&sym.FilterSoft,
			&sym.FilterSoftTicks,
			&sym.FilterHard,
			&sym.FilterHardTicks,
			&sym.FilterDiscard,
			&sym.FilterSpreadMax,
			&sym.FilterSpreadMin,
			&sym.SubscriptionsDelay,
			&sym.TradeMode,
			&sym.CalcMode,
			&sym.ExecMode,
			&sym.GtcMode,
			&sym.FillFlags,
			&sym.ExpirFlags,
			&sym.Spread,
			&sym.SpreadBalance,
			&sym.SpreadDiff,
			&sym.SpreadDiffBalance,
			&sym.TickValue,
			&sym.TickSize,
			&sym.ContractSize,
			&sym.StopsLevel,
			&sym.FreezeLevel,
			&sym.QuotesTimeout,
			&sym.VolumeMin,
			&sym.VolumeMinExt,
			&sym.VolumeMax,
			&sym.VolumeMaxExt,
			&sym.VolumeStep,
			&sym.VolumeStepExt,
			&sym.VolumeLimit,
			&sym.VolumeLimitExt,
			&sym.MarginFlags,
			&sym.MarginInitial,
			&sym.MarginMaintenance,
			&sym.MarginInitialBuy,
			&sym.MarginInitialSell,
			&sym.MarginInitialBuyLimit,
			&sym.MarginInitialSellLimit,
			&sym.MarginInitialBuyStop,
			&sym.MarginInitialSellStop,
			&sym.MarginInitialBuyStopLimit,
			&sym.MarginInitialSellStopLimit,
			&sym.MarginMaintenanceBuy,
			&sym.MarginMaintenanceSell,
			&sym.MarginMaintenanceBuyLimit,
			&sym.MarginMaintenanceSellLimit,
			&sym.MarginMaintenanceBuyStop,
			&sym.MarginMaintenanceSellStop,
			&sym.MarginMaintenanceBuyStopLimit,
			&sym.MarginMaintenanceSellStopLimit,
			&sym.MarginHedged,
			&sym.SwapMode,
			&sym.SwapLong,
			&sym.SwapShort,
			&sym.SwapYearDay,
			&sym.SwapFlags,
			&sym.SwapRateSunday,
			&sym.SwapRateMonday,
			&sym.SwapRateTuesday,
			&sym.SwapRateWednesday,
			&sym.SwapRateThursday,
			&sym.SwapRateFriday,
			&sym.SwapRateSaturday,
			&sym.TimeStart,
			&sym.TimeExpiration,
			&sym.ReFlags,
			&sym.ReTimeout,
			&sym.IeCheckMode,
			&sym.IeTimeout,
			&sym.IeSlipProfit,
			&sym.IeSlipLosing,
			&sym.IeVolumeMax,
			&sym.IeVolumeMaxExt,
			&sym.PriceSettle,
			&sym.PriceLimitMax,
			&sym.PriceLimitMin,
			&sym.TradeFlags,
			&sym.OrderFlags,
			&sym.MarginRateLiquidity,
			&sym.MarginRateCurrency,
			&sym.FaceValue,
			&sym.AccruedInterest,
			&sym.SpliceType,
			&sym.SpliceTimeType,
			&sym.SpliceTimeDays,
			&sym.OptionMode,
			&sym.PriceStrike,
			&sym.FilterGap,
			&sym.FilterGapTicks,
			&sym.TickChartMode,
			&sym.DateCreated,
			&sym.DateModified,
		)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT session_id, symbol_id, type, day, open, close
		   FROM hst.symbols_sessions
		  WHERE symbol_id = $1
		  ORDER BY type, day, open`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []model.SymbolSession{}
	for rows.Next() {
		var sess model.SymbolSession
		if err := rows.Scan(&sess.SessionId, &sess.SymbolId, &sess.Type, &sess.Day, &sess.Open, &sess.Close); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &ViewSymbolDetail{Symbol: sym, Sessions: sessions}, nil
}

// UpdateSymbol patches fields present in the body.
//
//	@Id			UpdateSymbol
//	@Tags		Symbols
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int			true	"symbol id"
//	@Param		body	body		UptSymbol	true	"only the fields to change"
//	@Success	200		{object}	Response{data=ViewSymbolDetail}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/{id} [patch]
func (s *Server) UpdateSymbol(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body UptSymbol
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}
	if body.Sessions != nil {
		if err := validateSessions(*body.Sessions); err != nil {
			return s.App.HttpResponseBadRequest(c, err)
		}
	}
	if body.VolumeMin != nil && body.VolumeMax != nil && *body.VolumeMax < *body.VolumeMin {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("volume_max must be >= volume_min"))
	}
	if body.VolumeStep != nil && *body.VolumeStep <= 0 {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("volume_step must be > 0"))
	}
	if body.Digits != nil && (*body.Digits < 0 || *body.Digits > 12) {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("digits must be 0..12"))
	}

	tx, err := s.DB.DB.Begin(ctx)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// one row as json, so the log can name what really changes without a wide scan
	var raw []byte
	if err := tx.QueryRow(ctx,
		`SELECT to_jsonb(s) FROM hst.symbols s WHERE symbol_id = $1 FOR UPDATE`, id).
		Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var before map[string]json.RawMessage
	if err := json.Unmarshal(raw, &before); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// only sessions need a second look, and only when the caller sends them
	var sessions []sessionWindow
	if body.Sessions != nil {
		rows, err := tx.Query(ctx,
			`SELECT type, day, open, close FROM hst.symbols_sessions WHERE symbol_id = $1`, id)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		for rows.Next() {
			var w sessionWindow
			if err := rows.Scan(&w.Type, &w.Day, &w.Open, &w.Close); err != nil {
				rows.Close()
				return s.App.HttpResponseInternalServerErrorRequest(c, err)
			}
			sessions = append(sessions, w)
		}
		rows.Close()
		if rows.Err() != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
		}
		sortSessions(sessions)
	}

	changes := changedSymbolFields(before, &body, sessions)

	// the stored path ends with the symbol, so renaming or moving rebuilds it
	if body.Symbol != nil || body.Path != nil {
		symbol := jsonString(before["symbol"])
		if body.Symbol != nil {
			symbol = *body.Symbol
		}
		folder := symbolFolder(jsonString(before["path"]))
		if body.Path != nil {
			folder = *body.Path
		}
		composed := symbolPath(folder, symbol)
		if len(composed) > 255 {
			return s.App.HttpResponseBadRequest(c, fmt.Errorf("path and symbol are longer than 255 together"))
		}
		body.Path = &composed
	}

	tag, err := tx.Exec(ctx,
		`UPDATE hst.symbols SET
		    symbol = COALESCE($2, symbol),
		    path = COALESCE($3, path),
		    isin = COALESCE($4, isin),
		    description = COALESCE($5, description),
		    international = COALESCE($6, international),
		    category = COALESCE($7, category),
		    exchange = COALESCE($8, exchange),
		    cfi = COALESCE($9, cfi),
		    sector = COALESCE($10, sector),
		    industry = COALESCE($11, industry),
		    country = COALESCE($12, country),
		    basis = COALESCE($13, basis),
		    source = COALESCE($14, source),
		    page = COALESCE($15, page),
		    currency_base = COALESCE($16, currency_base),
		    currency_base_digits = COALESCE($17, currency_base_digits),
		    currency_profit = COALESCE($18, currency_profit),
		    currency_profit_digits = COALESCE($19, currency_profit_digits),
		    currency_margin = COALESCE($20, currency_margin),
		    currency_margin_digits = COALESCE($21, currency_margin_digits),
		    color = COALESCE($22, color),
		    color_background = COALESCE($23, color_background),
		    digits = COALESCE($24, digits),
		    tick_flags = COALESCE($25, tick_flags),
		    tick_book_depth = COALESCE($26, tick_book_depth),
		    filter_soft = COALESCE($27, filter_soft),
		    filter_soft_ticks = COALESCE($28, filter_soft_ticks),
		    filter_hard = COALESCE($29, filter_hard),
		    filter_hard_ticks = COALESCE($30, filter_hard_ticks),
		    filter_discard = COALESCE($31, filter_discard),
		    filter_spread_max = COALESCE($32, filter_spread_max),
		    filter_spread_min = COALESCE($33, filter_spread_min),
		    subscriptions_delay = COALESCE($34, subscriptions_delay),
		    trade_mode = COALESCE($35, trade_mode),
		    calc_mode = COALESCE($36, calc_mode),
		    exec_mode = COALESCE($37, exec_mode),
		    gtc_mode = COALESCE($38, gtc_mode),
		    fill_flags = COALESCE($39, fill_flags),
		    expir_flags = COALESCE($40, expir_flags),
		    spread = COALESCE($41, spread),
		    spread_balance = COALESCE($42, spread_balance),
		    spread_diff = COALESCE($43, spread_diff),
		    spread_diff_balance = COALESCE($44, spread_diff_balance),
		    tick_value = COALESCE($45, tick_value),
		    tick_size = COALESCE($46, tick_size),
		    contract_size = COALESCE($47, contract_size),
		    stops_level = COALESCE($48, stops_level),
		    freeze_level = COALESCE($49, freeze_level),
		    quotes_timeout = COALESCE($50, quotes_timeout),
		    volume_min = COALESCE($51, volume_min),
		    volume_min_ext = COALESCE($52, volume_min_ext),
		    volume_max = COALESCE($53, volume_max),
		    volume_max_ext = COALESCE($54, volume_max_ext),
		    volume_step = COALESCE($55, volume_step),
		    volume_step_ext = COALESCE($56, volume_step_ext),
		    volume_limit = COALESCE($57, volume_limit),
		    volume_limit_ext = COALESCE($58, volume_limit_ext),
		    margin_flags = COALESCE($59, margin_flags),
		    margin_initial = COALESCE($60, margin_initial),
		    margin_maintenance = COALESCE($61, margin_maintenance),
		    margin_initial_buy = COALESCE($62, margin_initial_buy),
		    margin_initial_sell = COALESCE($63, margin_initial_sell),
		    margin_initial_buy_limit = COALESCE($64, margin_initial_buy_limit),
		    margin_initial_sell_limit = COALESCE($65, margin_initial_sell_limit),
		    margin_initial_buy_stop = COALESCE($66, margin_initial_buy_stop),
		    margin_initial_sell_stop = COALESCE($67, margin_initial_sell_stop),
		    margin_initial_buy_stop_limit = COALESCE($68, margin_initial_buy_stop_limit),
		    margin_initial_sell_stop_limit = COALESCE($69, margin_initial_sell_stop_limit),
		    margin_maintenance_buy = COALESCE($70, margin_maintenance_buy),
		    margin_maintenance_sell = COALESCE($71, margin_maintenance_sell),
		    margin_maintenance_buy_limit = COALESCE($72, margin_maintenance_buy_limit),
		    margin_maintenance_sell_limit = COALESCE($73, margin_maintenance_sell_limit),
		    margin_maintenance_buy_stop = COALESCE($74, margin_maintenance_buy_stop),
		    margin_maintenance_sell_stop = COALESCE($75, margin_maintenance_sell_stop),
		    margin_maintenance_buy_stop_limit = COALESCE($76, margin_maintenance_buy_stop_limit),
		    margin_maintenance_sell_stop_limit = COALESCE($77, margin_maintenance_sell_stop_limit),
		    margin_hedged = COALESCE($78, margin_hedged),
		    swap_mode = COALESCE($79, swap_mode),
		    swap_long = COALESCE($80, swap_long),
		    swap_short = COALESCE($81, swap_short),
		    swap_year_day = COALESCE($82, swap_year_day),
		    swap_flags = COALESCE($83, swap_flags),
		    swap_rate_sunday = COALESCE($84, swap_rate_sunday),
		    swap_rate_monday = COALESCE($85, swap_rate_monday),
		    swap_rate_tuesday = COALESCE($86, swap_rate_tuesday),
		    swap_rate_wednesday = COALESCE($87, swap_rate_wednesday),
		    swap_rate_thursday = COALESCE($88, swap_rate_thursday),
		    swap_rate_friday = COALESCE($89, swap_rate_friday),
		    swap_rate_saturday = COALESCE($90, swap_rate_saturday),
		    time_start = COALESCE($91, time_start),
		    time_expiration = COALESCE($92, time_expiration),
		    re_flags = COALESCE($93, re_flags),
		    re_timeout = COALESCE($94, re_timeout),
		    ie_check_mode = COALESCE($95, ie_check_mode),
		    ie_timeout = COALESCE($96, ie_timeout),
		    ie_slip_profit = COALESCE($97, ie_slip_profit),
		    ie_slip_losing = COALESCE($98, ie_slip_losing),
		    ie_volume_max = COALESCE($99, ie_volume_max),
		    ie_volume_max_ext = COALESCE($100, ie_volume_max_ext),
		    price_settle = COALESCE($101, price_settle),
		    price_limit_max = COALESCE($102, price_limit_max),
		    price_limit_min = COALESCE($103, price_limit_min),
		    trade_flags = COALESCE($104, trade_flags),
		    order_flags = COALESCE($105, order_flags),
		    margin_rate_liquidity = COALESCE($106, margin_rate_liquidity),
		    margin_rate_currency = COALESCE($107, margin_rate_currency),
		    face_value = COALESCE($108, face_value),
		    accrued_interest = COALESCE($109, accrued_interest),
		    splice_type = COALESCE($110, splice_type),
		    splice_time_type = COALESCE($111, splice_time_type),
		    splice_time_days = COALESCE($112, splice_time_days),
		    option_mode = COALESCE($113, option_mode),
		    price_strike = COALESCE($114, price_strike),
		    filter_gap = COALESCE($115, filter_gap),
		    filter_gap_ticks = COALESCE($116, filter_gap_ticks),
		    tick_chart_mode = COALESCE($117, tick_chart_mode),
		    point = CASE WHEN $24 IS NOT NULL THEN power(10::numeric, -($24)::int) ELSE point END,
		    multiply = CASE WHEN $24 IS NOT NULL THEN power(10::numeric, ($24)::int) ELSE multiply END,
		    date_modified = $118
		  WHERE symbol_id = $1`,
		id,
		body.Symbol,
		body.Path,
		body.Isin,
		body.Description,
		body.International,
		body.Category,
		body.Exchange,
		body.Cfi,
		body.Sector,
		body.Industry,
		body.Country,
		body.Basis,
		body.Source,
		body.Page,
		body.CurrencyBase,
		body.CurrencyBaseDigits,
		body.CurrencyProfit,
		body.CurrencyProfitDigits,
		body.CurrencyMargin,
		body.CurrencyMarginDigits,
		body.Color,
		body.ColorBackground,
		body.Digits,
		body.TickFlags,
		body.TickBookDepth,
		body.FilterSoft,
		body.FilterSoftTicks,
		body.FilterHard,
		body.FilterHardTicks,
		body.FilterDiscard,
		body.FilterSpreadMax,
		body.FilterSpreadMin,
		body.SubscriptionsDelay,
		body.TradeMode,
		body.CalcMode,
		body.ExecMode,
		body.GtcMode,
		body.FillFlags,
		body.ExpirFlags,
		body.Spread,
		body.SpreadBalance,
		body.SpreadDiff,
		body.SpreadDiffBalance,
		body.TickValue,
		body.TickSize,
		body.ContractSize,
		body.StopsLevel,
		body.FreezeLevel,
		body.QuotesTimeout,
		body.VolumeMin,
		body.VolumeMinExt,
		body.VolumeMax,
		body.VolumeMaxExt,
		body.VolumeStep,
		body.VolumeStepExt,
		body.VolumeLimit,
		body.VolumeLimitExt,
		body.MarginFlags,
		body.MarginInitial,
		body.MarginMaintenance,
		body.MarginInitialBuy,
		body.MarginInitialSell,
		body.MarginInitialBuyLimit,
		body.MarginInitialSellLimit,
		body.MarginInitialBuyStop,
		body.MarginInitialSellStop,
		body.MarginInitialBuyStopLimit,
		body.MarginInitialSellStopLimit,
		body.MarginMaintenanceBuy,
		body.MarginMaintenanceSell,
		body.MarginMaintenanceBuyLimit,
		body.MarginMaintenanceSellLimit,
		body.MarginMaintenanceBuyStop,
		body.MarginMaintenanceSellStop,
		body.MarginMaintenanceBuyStopLimit,
		body.MarginMaintenanceSellStopLimit,
		body.MarginHedged,
		body.SwapMode,
		body.SwapLong,
		body.SwapShort,
		body.SwapYearDay,
		body.SwapFlags,
		body.SwapRateSunday,
		body.SwapRateMonday,
		body.SwapRateTuesday,
		body.SwapRateWednesday,
		body.SwapRateThursday,
		body.SwapRateFriday,
		body.SwapRateSaturday,
		body.TimeStart,
		body.TimeExpiration,
		body.ReFlags,
		body.ReTimeout,
		body.IeCheckMode,
		body.IeTimeout,
		body.IeSlipProfit,
		body.IeSlipLosing,
		body.IeVolumeMax,
		body.IeVolumeMaxExt,
		body.PriceSettle,
		body.PriceLimitMax,
		body.PriceLimitMin,
		body.TradeFlags,
		body.OrderFlags,
		body.MarginRateLiquidity,
		body.MarginRateCurrency,
		body.FaceValue,
		body.AccruedInterest,
		body.SpliceType,
		body.SpliceTimeType,
		body.SpliceTimeDays,
		body.OptionMode,
		body.PriceStrike,
		body.FilterGap,
		body.FilterGapTicks,
		body.TickChartMode,
		time.Now().UnixNano())
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	if body.Sessions != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM hst.symbols_sessions WHERE symbol_id = $1`, id); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if err := insertSessions(ctx, tx, int64(id), *body.Sessions); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	detail, err := s.selectSymbolDetail(ctx, int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeOK, "symbol updated",
		"actor", snap.Login, "symbol_id", id,
		"symbol", detail.Symbol.Symbol, "changes", changes)

	if body.Sessions != nil {
		s.NotifyDatafeedsForSymbolID(c.UserContext(), int64(id))
	}
	s.NotifyWS(model.SubjectSymbol, model.EventSymbolUpdated, detail)
	s.NotifySystem(model.SubjectSystemSymbolUpdated, detail)
	s.JournalEntry(c, logger.CodeOK, journal.SymbolUpdatedMsg(snap.Login, detail.Symbol.Symbol), detail)

	return s.App.HttpResponseOK(c, detail)
}

// DeleteSymbol removes a symbol and its sessions.
//
//	@Id			DeleteSymbol
//	@Tags		Symbols
//	@Produce	json
//	@Param		id	path		int	true	"symbol id"
//	@Success	204	{object}	Response
//	@Failure	400	{object}	Response
//	@Failure	404	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/symbols/{id} [delete]
func (s *Server) DeleteSymbol(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var symbol, path string
	if err := s.DB.DB.QueryRow(c.UserContext(),
		`DELETE FROM hst.symbols WHERE symbol_id = $1 RETURNING symbol, path`, id).
		Scan(&symbol, &path); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	snap, _ := utils.GetClient(c)
	s.Log.Log(logger.TypeCfg, logger.CodeWarn, "symbol deleted",
		"actor", snap.Login, "symbol_id", id,
		"symbol", symbol, "path", path)

	ref := v1.ViewSymbolRef{SymbolId: id, Symbol: symbol, Path: path}
	s.NotifyWS(model.SubjectSymbol, model.EventSymbolDeleted, ref)
	s.NotifySystem(model.SubjectSystemSymbolDeleted, ref)
	s.JournalEntry(c, logger.CodeWarn, journal.SymbolDeletedMsg(snap.Login, symbol), ref)

	return s.App.HttpResponseNoContent(c)
}
