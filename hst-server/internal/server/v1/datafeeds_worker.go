package v1

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/events"
	"hstserver/pkg/logger"
	"hstserver/pkg/symbolpath"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

const workerDatafeedColumns = `datafeed_id, name, module, enable, mode,
	feed_server, feed_login, feed_password, timeout_reconnect`

const datafeedSymbolColumns = `feed_symbol_id, datafeed_id, symbol_id, path, exclude, symbol`

const symbolSettingsColumns = `symbol_id, path, digits, point, tick_flags, tick_book_depth,
	calc_mode, tick_chart_mode, splice_type,
	filter_soft, filter_soft_ticks, filter_hard, filter_hard_ticks, filter_discard,
	filter_spread_min, filter_spread_max, filter_gap, filter_gap_ticks,
	spread, spread_balance`

// ViewDatafeedQuoteSession is a quote session window for one feed-scoped symbol.
type ViewDatafeedQuoteSession struct {
	SymbolID int64  `json:"symbol_id"`
	Symbol   string `json:"symbol"`
	Day      int16  `json:"day"`
	Open     int32  `json:"open"`
	Close    int32  `json:"close"`
}

type datafeedScope struct {
	ResolvedSymbols      []ViewResolvedDatafeedSymbol
	QuoteSessions        []ViewDatafeedQuoteSession
	EffectiveSymbolCount int
	EffectiveTranslates  []events.WorkerTranslate
}

type feedSymbolRule struct {
	SymbolID int64
	Path     string
	Exclude  bool
}

func normalizeSymbolsMask(raw string) string {
	return strings.TrimSpace(raw)
}

func feedSymbolRulesFromViews(rows []ViewDatafeedSymbol) []feedSymbolRule {
	out := make([]feedSymbolRule, 0, len(rows))
	for _, row := range rows {
		r := feedSymbolRule{Path: row.Path, Exclude: row.Exclude}
		if row.SymbolID != nil {
			r.SymbolID = *row.SymbolID
		}
		out = append(out, r)
	}
	return out
}

func buildSelectedSymbolIDs(rules []feedSymbolRule, catalog []symbolpath.SymbolRef) map[int64]struct{} {
	selected := map[int64]struct{}{}
	for _, rule := range rules {
		if rule.SymbolID > 0 {
			symbolpath.ApplyScopeSymbol(selected, rule.SymbolID, rule.Exclude)
			continue
		}
		if rule.Path != "" {
			symbolpath.ApplyScopeRule(selected, rule.Exclude, rule.Path, catalog)
		}
	}
	return selected
}

func scanViewDatafeedSymbol(row pgx.Row) (*ViewDatafeedSymbol, error) {
	v := &ViewDatafeedSymbol{}
	var symbolID *int64
	var exclude int16
	err := row.Scan(&v.FeedSymbolID, &v.DatafeedID, &symbolID, &v.Path, &exclude, &v.Symbol)
	if err != nil {
		return nil, err
	}
	v.SymbolID = symbolID
	v.Exclude = exclude != 0
	return v, nil
}

func loadDatafeedSymbols(ctx context.Context, s *HttpServer, datafeedID int64) ([]ViewDatafeedSymbol, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+datafeedSymbolColumns+`
		   FROM hst.datafeed_symbols
		  WHERE datafeed_id = $1
		  ORDER BY feed_symbol_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewDatafeedSymbol{}
	for rows.Next() {
		v, err := scanViewDatafeedSymbol(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

func loadFeedSymbolRules(ctx context.Context, s *HttpServer, datafeedID int64) ([]feedSymbolRule, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT symbol_id, path, exclude
		   FROM hst.datafeed_symbols
		  WHERE datafeed_id = $1
		  ORDER BY feed_symbol_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []feedSymbolRule{}
	for rows.Next() {
		var symbolID *int64
		var path string
		var exclude int16
		if err := rows.Scan(&symbolID, &path, &exclude); err != nil {
			return nil, err
		}
		r := feedSymbolRule{Path: path, Exclude: exclude != 0}
		if symbolID != nil {
			r.SymbolID = *symbolID
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func loadFeedSymbolRulesByFeedIDs(ctx context.Context, s *HttpServer, ids []int64) (map[int64][]feedSymbolRule, error) {
	out := map[int64][]feedSymbolRule{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.DB.DB.Query(ctx,
		`SELECT datafeed_id, symbol_id, path, exclude
		   FROM hst.datafeed_symbols
		  WHERE datafeed_id = ANY($1)
		  ORDER BY datafeed_id, feed_symbol_id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feedID int64
		var symbolID *int64
		var path string
		var exclude int16
		if err := rows.Scan(&feedID, &symbolID, &path, &exclude); err != nil {
			return nil, err
		}
		r := feedSymbolRule{Path: path, Exclude: exclude != 0}
		if symbolID != nil {
			r.SymbolID = *symbolID
		}
		out[feedID] = append(out[feedID], r)
	}
	return out, rows.Err()
}

func (s *HttpServer) buildDatafeedScope(ctx context.Context, feedSymbols []ViewDatafeedSymbol, dbTranslates []ViewDatafeedTranslate) (*datafeedScope, error) {
	catalog, err := loadCatalogSymbolRefs(ctx, s)
	if err != nil {
		return nil, err
	}

	dbWorker := make([]events.WorkerTranslate, 0, len(dbTranslates))
	for _, tr := range dbTranslates {
		dbWorker = append(dbWorker, events.WorkerTranslate{
			TranslateID: tr.TranslateID,
			SymbolID:    tr.SymbolID,
			Symbol:      tr.Symbol,
			Source:      tr.Source,
			BidMarkup:   tr.BidMarkup,
			AskMarkup:   tr.AskMarkup,
			Digits:      tr.Digits,
		})
	}

	effective := mergeEffectiveTranslates(dbWorker, catalog, feedSymbolRulesFromViews(feedSymbols))
	byCatalog := make(map[int64]symbolpath.SymbolRef, len(catalog))
	for _, sym := range catalog {
		byCatalog[sym.SymbolID] = sym
	}

	resolved := make([]ViewResolvedDatafeedSymbol, 0, len(effective))
	for _, tr := range effective {
		sym, ok := byCatalog[tr.SymbolID]
		if !ok {
			continue
		}
		resolved = append(resolved, ViewResolvedDatafeedSymbol{
			SymbolID: sym.SymbolID,
			Symbol:   sym.Symbol,
			Path:     sym.Path,
		})
	}
	sort.Slice(resolved, func(i, j int) bool { return resolved[i].SymbolID < resolved[j].SymbolID })

	symbolIDs := translateSymbolIDs(effective)
	sessionsBySymbol, err := loadQuoteSessionsBySymbolIDs(ctx, s, symbolIDs)
	if err != nil {
		return nil, err
	}

	quoteSessions := make([]ViewDatafeedQuoteSession, 0)
	for _, tr := range effective {
		for _, sess := range sessionsBySymbol[tr.SymbolID] {
			quoteSessions = append(quoteSessions, ViewDatafeedQuoteSession{
				SymbolID: tr.SymbolID,
				Symbol:   tr.Symbol,
				Day:      sess.Day,
				Open:     sess.Open,
				Close:    sess.Close,
			})
		}
	}

	return &datafeedScope{
		ResolvedSymbols:      resolved,
		QuoteSessions:        quoteSessions,
		EffectiveSymbolCount: len(effective),
		EffectiveTranslates:  effective,
	}, nil
}

type workerRowScanner interface {
	Scan(dest ...any) error
}

func scanWorkerDatafeed(row workerRowScanner) (*events.WorkerDatafeedConfig, error) {
	cfg := &events.WorkerDatafeedConfig{}
	err := row.Scan(
		&cfg.DatafeedID, &cfg.Name, &cfg.Module, &cfg.Enable, &cfg.Mode,
		&cfg.FeedServer, &cfg.FeedLogin, &cfg.FeedPassword, &cfg.TimeoutReconnect,
	)
	return cfg, err
}

func (s *HttpServer) loadWorkerDatafeedConfig(ctx context.Context, datafeedID int64) (*events.WorkerDatafeedConfig, error) {
	cfg, err := scanWorkerDatafeed(s.DB.DB.QueryRow(ctx,
		`SELECT `+workerDatafeedColumns+` FROM hst.datafeeds WHERE datafeed_id = $1`, datafeedID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	params, err := loadDatafeedParams(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}
	cfg.Params = make([]events.WorkerParam, 0, len(params))
	for _, p := range params {
		cfg.Params = append(cfg.Params, events.WorkerParam{
			ParamKey: p.ParamKey,
			Value:    p.Value,
		})
	}

	catalog, err := loadCatalogSymbolRefs(ctx, s)
	if err != nil {
		return nil, err
	}

	dbTranslates, err := loadDBWorkerTranslates(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}
	rules, err := loadFeedSymbolRules(ctx, s, datafeedID)
	if err != nil {
		return nil, err
	}
	cfg.Translates = mergeEffectiveTranslates(dbTranslates, catalog, rules)

	symbolIDs := translateSymbolIDs(cfg.Translates)
	cfg.Sessions, err = loadQuoteSessionsForSymbols(ctx, s, symbolIDs)
	if err != nil {
		return nil, err
	}

	cfg.Settings, err = loadSymbolSettingsForSymbols(ctx, s, symbolIDs)
	if err != nil {
		return nil, err
	}

	cfg.Holidays, err = loadWorkerHolidays(ctx, s)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *HttpServer) listWorkerDatafeedConfigs(ctx context.Context, mode int32) ([]events.WorkerDatafeedConfig, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+workerDatafeedColumns+`
		   FROM hst.datafeeds
		  WHERE enable = 1
		    AND ($1 = 0 OR (mode & $1) = $1)
		  ORDER BY datafeed_id`, mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type feedRow struct {
		cfg events.WorkerDatafeedConfig
	}
	var feedRows []feedRow
	var ids []int64
	for rows.Next() {
		cfg, err := scanWorkerDatafeed(rows)
		if err != nil {
			return nil, err
		}
		feedRows = append(feedRows, feedRow{cfg: *cfg})
		ids = append(ids, cfg.DatafeedID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []events.WorkerDatafeedConfig{}, nil
	}

	paramRows, err := s.DB.DB.Query(ctx,
		`SELECT datafeed_id, param_key, value
		   FROM hst.datafeed_params
		  WHERE datafeed_id = ANY($1)
		  ORDER BY datafeed_id, priority, param_id`, ids)
	if err != nil {
		return nil, err
	}
	defer paramRows.Close()

	paramsByFeed := map[int64][]events.WorkerParam{}
	for paramRows.Next() {
		var feedID int64
		var p events.WorkerParam
		if err := paramRows.Scan(&feedID, &p.ParamKey, &p.Value); err != nil {
			return nil, err
		}
		paramsByFeed[feedID] = append(paramsByFeed[feedID], p)
	}
	if err := paramRows.Err(); err != nil {
		return nil, err
	}

	trByFeed, err := loadDBWorkerTranslatesByFeedIDs(ctx, s, ids)
	if err != nil {
		return nil, err
	}

	rulesByFeed, err := loadFeedSymbolRulesByFeedIDs(ctx, s, ids)
	if err != nil {
		return nil, err
	}

	catalog, err := loadCatalogSymbolRefs(ctx, s)
	if err != nil {
		return nil, err
	}

	out := make([]events.WorkerDatafeedConfig, 0, len(feedRows))
	allSymbolIDs := map[int64]struct{}{}
	for _, row := range feedRows {
		cfg := row.cfg
		cfg.Params = paramsByFeed[cfg.DatafeedID]
		cfg.Translates = mergeEffectiveTranslates(trByFeed[cfg.DatafeedID], catalog, rulesByFeed[cfg.DatafeedID])
		for _, id := range translateSymbolIDs(cfg.Translates) {
			allSymbolIDs[id] = struct{}{}
		}
		out = append(out, cfg)
	}

	sessionIDs := make([]int64, 0, len(allSymbolIDs))
	for id := range allSymbolIDs {
		sessionIDs = append(sessionIDs, id)
	}
	sessionsBySymbol, err := loadQuoteSessionsBySymbolIDs(ctx, s, sessionIDs)
	if err != nil {
		return nil, err
	}
	settingsBySymbol, err := loadSymbolSettingsBySymbolIDs(ctx, s, sessionIDs)
	if err != nil {
		return nil, err
	}
	holidays, err := loadWorkerHolidays(ctx, s)
	if err != nil {
		return nil, err
	}

	for i := range out {
		out[i].Sessions = flattenQuoteSessions(out[i].Translates, sessionsBySymbol)
		out[i].Settings = symbolSettingsForTranslates(out[i].Translates, settingsBySymbol)
		out[i].Holidays = holidays
	}

	return out, nil
}

func (s *HttpServer) publishWorkerConfigSnapshot(ctx context.Context, datafeedID int64) {
	cfg, err := s.loadWorkerDatafeedConfig(ctx, datafeedID)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			s.Log.Log(logger.TypeNet, logger.CodeWarn, "worker config load failed",
				"datafeed_id", datafeedID, "error", err.Error())
		}
		return
	}
	if err := events.PublishConfigSnapshot(s.Nats, *cfg); err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "worker config publish failed",
			"datafeed_id", datafeedID, "error", err.Error())
	}
}

// NotifyAllDatafeedConfigs republishes every enabled feed's snapshot. The holiday calendar is
// server-wide and its masks can reach any symbol, so there is no subset worth working out.
func (s *HttpServer) NotifyAllDatafeedConfigs(ctx context.Context) {
	rows, err := s.DB.DB.Query(ctx, `SELECT datafeed_id FROM hst.datafeeds WHERE enable = 1`)
	if err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "datafeed list failed for calendar notify",
			"error", err.Error())
		return
	}

	// drain first: publishing borrows more connections, and this cursor is holding one
	var feedIDs []int64
	for rows.Next() {
		var feedID int64
		if err := rows.Scan(&feedID); err != nil {
			rows.Close()
			return
		}
		feedIDs = append(feedIDs, feedID)
	}
	rows.Close()

	for _, feedID := range feedIDs {
		s.publishWorkerConfigSnapshot(ctx, feedID)
	}
}

func (s *HttpServer) NotifyDatafeedsForSymbolID(ctx context.Context, symbolID int64) {
	catalog, err := loadCatalogSymbolRefs(ctx, s)
	if err != nil {
		s.Log.Log(logger.TypeNet, logger.CodeWarn, "catalog load failed for session notify",
			"symbol_id", symbolID, "error", err.Error())
		return
	}

	var symRef *symbolpath.SymbolRef
	for i := range catalog {
		if catalog[i].SymbolID == symbolID {
			symRef = &catalog[i]
			break
		}
	}
	if symRef == nil {
		return
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT DISTINCT datafeed_id FROM hst.datafeeds WHERE enable = 1`)
	if err != nil {
		return
	}
	// drain first: the checks below borrow more connections, and this cursor is holding one
	var feedIDs []int64
	for rows.Next() {
		var feedID int64
		if err := rows.Scan(&feedID); err != nil {
			rows.Close()
			return
		}
		feedIDs = append(feedIDs, feedID)
	}
	rows.Close()

	for _, feedID := range feedIDs {
		if !feedIncludesSymbol(ctx, s, feedID, *symRef, catalog) {
			continue
		}
		// only the config snapshot: a symbol edit hot-applies, it must not bounce the live feed
		s.publishWorkerConfigSnapshot(ctx, feedID)
	}
}

func feedIncludesSymbol(ctx context.Context, s *HttpServer, feedID int64, sym symbolpath.SymbolRef, catalog []symbolpath.SymbolRef) bool {
	rules, err := loadFeedSymbolRules(ctx, s, feedID)
	if err != nil {
		return false
	}
	selected := buildSelectedSymbolIDs(rules, catalog)
	if _, ok := selected[sym.SymbolID]; ok {
		return true
	}
	if len(rules) > 0 {
		return false
	}
	var count int
	if err := s.DB.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM hst.datafeed_translates WHERE datafeed_id = $1 AND symbol_id = $2`,
		feedID, sym.SymbolID).Scan(&count); err == nil && count > 0 {
		return true
	}
	return false
}

// ListInternalDatafeeds returns enabled feeds for worker services.
func (s *HttpServer) ListInternalDatafeeds(c *fiber.Ctx) error {
	mode := int32(0)
	if raw := c.Query("mode"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return s.App.HttpResponseBadQueryParams(c, err)
		}
		mode = int32(v)
		if mode != 0 {
			if err := validateFeederMode(model.FeederFlags(mode)); err != nil {
				return s.App.HttpResponseBadRequest(c, err)
			}
		}
	}

	out, err := s.listWorkerDatafeedConfigs(c.UserContext(), mode)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, out)
}

// GetInternalDatafeed returns one feed config for worker services.
func (s *HttpServer) GetInternalDatafeed(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	cfg, err := s.loadWorkerDatafeedConfig(c.UserContext(), int64(id))
	if errors.Is(err, errs.ErrNotFound) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	return s.App.HttpResponseOK(c, cfg)
}

func loadCatalogSymbolRefs(ctx context.Context, s *HttpServer) ([]symbolpath.SymbolRef, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT symbol_id, symbol, path FROM hst.symbols ORDER BY symbol_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []symbolpath.SymbolRef{}
	for rows.Next() {
		var sym symbolpath.SymbolRef
		if err := rows.Scan(&sym.SymbolID, &sym.Symbol, &sym.Path); err != nil {
			return nil, err
		}
		out = append(out, sym)
	}
	return out, rows.Err()
}

func loadDBWorkerTranslates(ctx context.Context, s *HttpServer, datafeedID int64) ([]events.WorkerTranslate, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT translate_id, symbol_id, symbol, source, bid_markup, ask_markup, digits
		   FROM hst.datafeed_translates
		  WHERE datafeed_id = $1
		  ORDER BY translate_id`, datafeedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []events.WorkerTranslate{}
	for rows.Next() {
		var t events.WorkerTranslate
		if err := rows.Scan(&t.TranslateID, &t.SymbolID, &t.Symbol, &t.Source,
			&t.BidMarkup, &t.AskMarkup, &t.Digits); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func loadDBWorkerTranslatesByFeedIDs(ctx context.Context, s *HttpServer, ids []int64) (map[int64][]events.WorkerTranslate, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT datafeed_id, translate_id, symbol_id, symbol, source,
		        bid_markup, ask_markup, digits
		   FROM hst.datafeed_translates
		  WHERE datafeed_id = ANY($1)
		  ORDER BY datafeed_id, translate_id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64][]events.WorkerTranslate{}
	for rows.Next() {
		var feedID int64
		var t events.WorkerTranslate
		if err := rows.Scan(&feedID, &t.TranslateID, &t.SymbolID, &t.Symbol, &t.Source,
			&t.BidMarkup, &t.AskMarkup, &t.Digits); err != nil {
			return nil, err
		}
		out[feedID] = append(out[feedID], t)
	}
	return out, rows.Err()
}

func mergeEffectiveTranslates(db []events.WorkerTranslate, catalog []symbolpath.SymbolRef, rules []feedSymbolRule) []events.WorkerTranslate {
	hasScopeRules := len(rules) > 0
	selected := buildSelectedSymbolIDs(rules, catalog)

	// Explicit translate rows always belong to feed scope (MT5: translation implies symbol).
	for _, tr := range db {
		if tr.SymbolID > 0 {
			selected[tr.SymbolID] = struct{}{}
		}
	}

	byCatalog := make(map[int64]symbolpath.SymbolRef, len(catalog))
	for _, sym := range catalog {
		byCatalog[sym.SymbolID] = sym
	}

	byID := make(map[int64]events.WorkerTranslate, len(db))
	for _, tr := range db {
		if tr.SymbolID <= 0 {
			continue
		}
		if hasScopeRules {
			if _, ok := selected[tr.SymbolID]; !ok {
				continue
			}
		}
		byID[tr.SymbolID] = tr
	}

	if hasScopeRules {
		for id := range selected {
			if _, ok := byID[id]; ok {
				continue
			}
			sym, ok := byCatalog[id]
			if !ok {
				continue
			}
			byID[id] = events.WorkerTranslate{
				SymbolID: id,
				Symbol:   sym.Symbol,
				Source:   "",
			}
		}
	}

	order := make([]int64, 0, len(byID))
	for id := range byID {
		order = append(order, id)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })

	out := make([]events.WorkerTranslate, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

func translateSymbolIDs(translates []events.WorkerTranslate) []int64 {
	out := make([]int64, 0, len(translates))
	for _, tr := range translates {
		if tr.SymbolID > 0 {
			out = append(out, tr.SymbolID)
		}
	}
	return out
}

func loadQuoteSessionsForSymbols(ctx context.Context, s *HttpServer, symbolIDs []int64) ([]events.WorkerSymbolSession, error) {
	bySymbol, err := loadQuoteSessionsBySymbolIDs(ctx, s, symbolIDs)
	if err != nil {
		return nil, err
	}
	return flattenQuoteSessionsFromIDs(symbolIDs, bySymbol), nil
}

func flattenQuoteSessions(translates []events.WorkerTranslate, bySymbol map[int64][]events.WorkerSymbolSession) []events.WorkerSymbolSession {
	return flattenQuoteSessionsFromIDs(translateSymbolIDs(translates), bySymbol)
}

func flattenQuoteSessionsFromIDs(symbolIDs []int64, bySymbol map[int64][]events.WorkerSymbolSession) []events.WorkerSymbolSession {
	out := []events.WorkerSymbolSession{}
	for _, id := range symbolIDs {
		out = append(out, bySymbol[id]...)
	}
	return out
}

func loadSymbolSettingsForSymbols(ctx context.Context, s *HttpServer, symbolIDs []int64) ([]events.WorkerSymbolSettings, error) {
	bySymbol, err := loadSymbolSettingsBySymbolIDs(ctx, s, symbolIDs)
	if err != nil {
		return nil, err
	}
	return flattenSymbolSettingsFromIDs(symbolIDs, bySymbol), nil
}

func symbolSettingsForTranslates(translates []events.WorkerTranslate, bySymbol map[int64]events.WorkerSymbolSettings) []events.WorkerSymbolSettings {
	return flattenSymbolSettingsFromIDs(translateSymbolIDs(translates), bySymbol)
}

func flattenSymbolSettingsFromIDs(symbolIDs []int64, bySymbol map[int64]events.WorkerSymbolSettings) []events.WorkerSymbolSettings {
	out := []events.WorkerSymbolSettings{}
	for _, id := range symbolIDs {
		if set, ok := bySymbol[id]; ok {
			out = append(out, set)
		}
	}
	return out
}

func loadSymbolSettingsBySymbolIDs(ctx context.Context, s *HttpServer, symbolIDs []int64) (map[int64]events.WorkerSymbolSettings, error) {
	out := map[int64]events.WorkerSymbolSettings{}
	if len(symbolIDs) == 0 {
		return out, nil
	}
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+symbolSettingsColumns+`
		   FROM hst.symbols
		  WHERE symbol_id = ANY($1)
		  ORDER BY symbol_id`, symbolIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var set events.WorkerSymbolSettings
		if err := rows.Scan(&set.SymbolID, &set.Path, &set.Digits, &set.Point, &set.TickFlags, &set.TickBookDepth,
			&set.CalcMode, &set.TickChartMode, &set.SpliceType,
			&set.FilterSoft, &set.FilterSoftTicks, &set.FilterHard, &set.FilterHardTicks, &set.FilterDiscard,
			&set.FilterSpreadMin, &set.FilterSpreadMax, &set.FilterGap, &set.FilterGapTicks,
			&set.Spread, &set.SpreadBalance); err != nil {
			return nil, err
		}
		out[set.SymbolID] = set
	}
	return out, rows.Err()
}

// loadWorkerHolidays reads the whole enabled calendar. It is a handful of rows and the masks
// are resolved by the worker against each symbol's path, so there is nothing to filter here.
func loadWorkerHolidays(ctx context.Context, s *HttpServer) ([]events.WorkerHoliday, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT year, month, day, "from", "to", symbols
		   FROM hst.holidays
		  WHERE mode = 1
		  ORDER BY config_index`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []events.WorkerHoliday{}
	for rows.Next() {
		var h events.WorkerHoliday
		if err := rows.Scan(&h.Year, &h.Month, &h.Day, &h.From, &h.To, &h.Symbols); err != nil {
			return nil, err
		}
		out = append(out, h)
	}

	return out, rows.Err()
}

func loadQuoteSessionsBySymbolIDs(ctx context.Context, s *HttpServer, symbolIDs []int64) (map[int64][]events.WorkerSymbolSession, error) {
	out := map[int64][]events.WorkerSymbolSession{}
	if len(symbolIDs) == 0 {
		return out, nil
	}
	rows, err := s.DB.DB.Query(ctx,
		`SELECT symbol_id, day, open, close
		   FROM hst.symbols_sessions
		  WHERE symbol_id = ANY($1)
		    AND type = $2
		  ORDER BY symbol_id, day, open`, symbolIDs, model.SymbolSessionType_quote)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sess events.WorkerSymbolSession
		if err := rows.Scan(&sess.SymbolID, &sess.Day, &sess.Open, &sess.Close); err != nil {
			return nil, err
		}
		out[sess.SymbolID] = append(out[sess.SymbolID], sess)
	}
	return out, rows.Err()
}
