package trader

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ViewSymbolNode is one folder or one instrument in the Market Watch browser tree.
type ViewSymbolNode struct {
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	IsFolder bool              `json:"is_folder"`
	Children []*ViewSymbolNode `json:"children,omitempty"`
	Symbol   *model.SymbolInfo `json:"symbol,omitempty"`
}

// ViewSessionWindow is one open window inside a weekday, in minutes since midnight.
type ViewSessionWindow struct {
	Open  int32  `json:"open"`
	Close int32  `json:"close"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// ViewSymbolDay is a weekday's quote and trade windows; empty lists mean closed all day.
type ViewSymbolDay struct {
	Day   int16               `json:"day"`
	Quote []ViewSessionWindow `json:"quote"`
	Trade []ViewSessionWindow `json:"trade"`
}

// ViewSymbolSessions is the whole week for one symbol, Sunday first, every day present.
type ViewSymbolSessions struct {
	Symbol string          `json:"symbol"`
	Days   []ViewSymbolDay `json:"days"`
	// Holiday is set when today is a holiday for this symbol: closed all day when Windows is
	// empty, otherwise open only inside them. Times are server time.
	Holiday *ViewSymbolHoliday `json:"holiday,omitempty"`
	// Leverage is the group's floating leverage rule covering this symbol, when it has one.
	Leverage *ViewSymbolLeverage `json:"leverage,omitempty"`
}

// ViewSymbolHoliday is today's holiday as the sessions dialog shows it.
type ViewSymbolHoliday struct {
	Windows     []ViewSessionWindow `json:"windows"`
	Description string              `json:"description"`
}

// ViewSymbolLeverage is the floating leverage the caller's group applies to this symbol:
// the first rule whose path covers it, tier by tier.
type ViewSymbolLeverage struct {
	Name      string             `json:"name"`
	RangeMode int32              `json:"range_mode"`
	Currency  string             `json:"currency"`
	Tiers     []ViewLeverageTier `json:"tiers"`
}

// ViewLeverageTier is one level; RangeTo 0 on the last tier means no upper bound.
type ViewLeverageTier struct {
	RangeFrom             float64 `json:"range_from"`
	RangeTo               float64 `json:"range_to"`
	MarginRateInitial     float64 `json:"margin_rate_initial"`
	MarginRateMaintenance float64 `json:"margin_rate_maintenance"`
}

// MySymbolsTree returns the caller's symbols nested by their backslash separated path.
//
//	@Id			MySymbolsTree
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewSymbolNode}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/symbols/tree [get]
func (s *Server) MySymbolsTree(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	list, err := s.resolveMySymbols(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, symbolTree(list))
}

// MySymbolByName returns one of the caller's symbols, in the same shape the list gives.
//
//	@Id			MySymbolByName
//	@Tags		Trader
//	@Produce	json
//	@Param		symbol	query		string	true	"symbol name"
//	@Success	200		{object}	Response{data=model.SymbolInfo}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/symbols/by_name [get]
func (s *Server) MySymbolByName(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	name := strings.TrimSpace(c.Query("symbol"))
	if name == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	list, err := s.resolveMySymbols(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	for i := range list {
		if strings.EqualFold(list[i].Symbol, name) {
			return s.App.HttpResponseOK(c, list[i])
		}
	}

	return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
}

// MySymbolSessions returns the weekly quote and trade schedule for one of the caller's symbols.
//
//	@Id			MySymbolSessions
//	@Tags		Trader
//	@Produce	json
//	@Param		symbol	path		string	true	"symbol name"
//	@Success	200		{object}	Response{data=ViewSymbolSessions}
//	@Failure	403		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/symbols/{symbol}/sessions [get]
func (s *Server) MySymbolSessions(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	name := strings.TrimSpace(c.Params("symbol"))
	if name == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	list, err := s.resolveMySymbols(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// a symbol the group never granted must not leak its schedule either
	granted, path := "", ""
	for i := range list {
		if strings.EqualFold(list[i].Symbol, name) {
			granted, path = list[i].Symbol, list[i].Path
			break
		}
	}
	if granted == "" {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT ss.type, ss.day, ss.open, ss.close
		   FROM hst.symbols_sessions ss
		   JOIN hst.symbols s ON s.symbol_id = ss.symbol_id
		  WHERE s.symbol = $1
		  ORDER BY ss.day, ss.open`, granted)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := ViewSymbolSessions{Symbol: granted, Days: make([]ViewSymbolDay, 7)}
	for i := range out.Days {
		out.Days[i] = ViewSymbolDay{Day: int16(i), Quote: []ViewSessionWindow{}, Trade: []ViewSessionWindow{}}
	}

	for rows.Next() {
		var kind, day int16
		var w ViewSessionWindow
		if err := rows.Scan(&kind, &day, &w.Open, &w.Close); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if day < 0 || day > 6 {
			continue
		}
		w.From, w.To = hhmm(w.Open), hhmm(w.Close)
		if kind == 0 {
			out.Days[day].Quote = append(out.Days[day].Quote, w)
		} else {
			out.Days[day].Trade = append(out.Days[day].Trade, w)
		}
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	holiday, err := s.todaysHoliday(c.UserContext(), path, granted)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	out.Holiday = holiday

	leverage, err := s.symbolLeverage(c.UserContext(), snap.Login, path, granted)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	out.Leverage = leverage

	return s.App.HttpResponseOK(c, out)
}

// symbolLeverage resolves the caller's group profile and returns the first rule that covers the
// symbol, the same way the engine picks its margin rate.
func (s *Server) symbolLeverage(ctx context.Context, login int64, path, symbol string) (*ViewSymbolLeverage, error) {
	var leverageId *int64
	err := s.DB.DB.QueryRow(ctx,
		`SELECT g.margin_leverage_id
		   FROM hst.users u
		   JOIN hst.groups g ON g."group" = u."group"
		  WHERE u.login = $1`, login).Scan(&leverageId)
	if err != nil || leverageId == nil {
		return nil, nil
	}

	rows, err := s.DB.DB.Query(ctx,
		`SELECT rule_id, name, path, range_mode, range_value_currency
		   FROM hst.leverage_rules
		  WHERE leverage_id = $1
		  ORDER BY config_index`, *leverageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out *ViewSymbolLeverage
	var ruleId int64
	for rows.Next() {
		var id int64
		var name, mask, currency string
		var mode int32
		if err := rows.Scan(&id, &name, &mask, &mode, &currency); err != nil {
			return nil, err
		}
		if out == nil && HolidayLayer([]string{mask}, path, symbol) {
			out = &ViewSymbolLeverage{Name: name, RangeMode: mode, Currency: currency, Tiers: []ViewLeverageTier{}}
			ruleId = id
		}
	}
	if rows.Err() != nil || out == nil {
		return out, rows.Err()
	}

	tiers, err := s.DB.DB.Query(ctx,
		`SELECT range_from, range_to, margin_rate_initial, margin_rate_maintenance
		   FROM hst.leverage_tiers
		  WHERE rule_id = $1
		  ORDER BY range_from`, ruleId)
	if err != nil {
		return nil, err
	}
	defer tiers.Close()

	for tiers.Next() {
		var t ViewLeverageTier
		if err := tiers.Scan(&t.RangeFrom, &t.RangeTo, &t.MarginRateInitial, &t.MarginRateMaintenance); err != nil {
			return nil, err
		}
		out.Tiers = append(out.Tiers, t)
	}

	return out, tiers.Err()
}

// todaysHoliday reads whether today, in server time, is a holiday for this symbol; the engine
// matches the same rows in its own calendar.
func (s *Server) todaysHoliday(ctx context.Context, path, symbol string) (*ViewSymbolHoliday, error) {
	now := time.Now().UTC()

	rows, err := s.DB.DB.Query(ctx,
		`SELECT "from", "to", symbols, description
		   FROM hst.holidays
		  WHERE mode = 1
		    AND (year = 0 OR year = $1)
		    AND month = $2 AND day = $3
		  ORDER BY config_index`, now.Year(), int(now.Month()), now.Day())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holiday *ViewSymbolHoliday
	for rows.Next() {
		var from, to int32
		var masks []string
		var description string
		if err := rows.Scan(&from, &to, &masks, &description); err != nil {
			return nil, err
		}
		if !HolidayLayer(masks, path, symbol) {
			continue
		}
		if holiday == nil {
			holiday = &ViewSymbolHoliday{Windows: []ViewSessionWindow{}}
		}
		if from != 0 || to != 0 {
			holiday.Windows = append(holiday.Windows,
				ViewSessionWindow{Open: from, Close: to, From: hhmm(from), To: hhmm(to)})
		}
		if description != "" {
			if holiday.Description != "" {
				holiday.Description += "; "
			}
			holiday.Description += description
		}
	}

	return holiday, rows.Err()
}

// HolidayLayer reports whether a holiday's symbol masks cover this symbol: exact name or path,
// or a trailing-star prefix of either.
func HolidayLayer(masks []string, path, symbol string) bool {
	if len(masks) == 0 {
		return true
	}
	for _, raw := range masks {
		m := strings.TrimSpace(raw)
		switch {
		case m == "" || m == "*":
			return true
		case strings.HasSuffix(m, "*"):
			prefix := strings.TrimSuffix(m, "*")
			if strings.HasPrefix(path, prefix) || strings.HasPrefix(symbol, prefix) {
				return true
			}
		case m == symbol || m == path:
			return true
		}
	}
	return false
}

// hhmm renders minutes since midnight, where 1440 is the end of the day.
func hhmm(m int32) string {
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

// symbolTree nests the flat list on its backslash separated path, folders before leaves.
func symbolTree(list []model.SymbolInfo) []*ViewSymbolNode {
	root := &ViewSymbolNode{IsFolder: true}
	folders := map[string]*ViewSymbolNode{"": root}

	for i := range list {
		// the stored path ends with the symbol itself, so the folders are everything before it
		parts := strings.Split(strings.Trim(list[i].Path, `\`), `\`)
		if n := len(parts); n > 0 && strings.EqualFold(parts[n-1], list[i].Symbol) {
			parts = parts[:n-1]
		}

		parent, prefix := root, ""
		for _, name := range parts {
			if name == "" {
				continue
			}
			if prefix == "" {
				prefix = name
			} else {
				prefix += `\` + name
			}
			node, ok := folders[prefix]
			if !ok {
				node = &ViewSymbolNode{Name: name, Path: prefix, IsFolder: true}
				folders[prefix] = node
				parent.Children = append(parent.Children, node)
			}
			parent = node
		}

		parent.Children = append(parent.Children, &ViewSymbolNode{
			Name:   list[i].Symbol,
			Path:   list[i].Path,
			Symbol: &list[i],
		})
	}

	sortNodes(root)
	if root.Children == nil {
		return []*ViewSymbolNode{}
	}
	return root.Children
}

func sortNodes(n *ViewSymbolNode) {
	sort.SliceStable(n.Children, func(i, j int) bool {
		a, b := n.Children[i], n.Children[j]
		if a.IsFolder != b.IsFolder {
			return a.IsFolder
		}
		return a.Name < b.Name
	})
	for _, c := range n.Children {
		if c.IsFolder {
			sortNodes(c)
		}
	}
}

// resolveMySymbols is the one place a trader's symbol list comes from, engine first then database.
func (s *Server) resolveMySymbols(ctx context.Context, login int64) ([]model.SymbolInfo, error) {
	if live := s.AskEngineSymbols(ctx, login); len(live) > 0 {
		return live, nil
	}

	// the group's overrides decide what this trader may see, so a symbol nobody granted it never
	// appears; the first override row in config order wins, an absent field inherits the symbol
	rows, err := s.DB.DB.Query(ctx,
		`SELECT s.symbol, s.path, s.description, s.digits,
		        COALESCE(o.trade_mode, s.trade_mode),
		        s.calc_mode,
		        COALESCE(o.exec_mode, s.exec_mode),
		        COALESCE(o.stops_level, s.stops_level),
		        COALESCE(NULLIF(COALESCE(o.volume_min_ext, s.volume_min_ext), 0) / 100000000.0,
		                 COALESCE(o.volume_min, s.volume_min) / 10000.0),
		        COALESCE(NULLIF(COALESCE(o.volume_max_ext, s.volume_max_ext), 0) / 100000000.0,
		                 COALESCE(o.volume_max, s.volume_max) / 10000.0),
		        COALESCE(NULLIF(COALESCE(o.volume_step_ext, s.volume_step_ext), 0) / 100000000.0,
		                 COALESCE(o.volume_step, s.volume_step) / 10000.0),
		        s.contract_size
		   FROM hst.symbols s
		   JOIN hst.users u ON u.login = $1
		   JOIN hst.groups g ON g."group" = u."group"
		   JOIN LATERAL (
		        SELECT gs.trade_mode, gs.exec_mode,
		               gs.stops_level, gs.volume_min, gs.volume_max, gs.volume_step,
		               gs.volume_min_ext, gs.volume_max_ext, gs.volume_step_ext
		          FROM hst.groups_symbols gs
		         WHERE gs.group_id = g.group_id
		           AND (gs.path = '*' OR s.path = gs.path OR starts_with(s.path, rtrim(gs.path, '*')))
		         ORDER BY gs.config_index
		         LIMIT 1) o ON TRUE
		  ORDER BY s.symbol`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.SymbolInfo{}
	for rows.Next() {
		var v model.SymbolInfo
		if err := rows.Scan(&v.Symbol, &v.Path, &v.Description, &v.Digits,
			&v.TradeMode, &v.CalcMode, &v.ExecMode,
			&v.StopsLevel, &v.VolumeMin, &v.VolumeMax, &v.VolumeStep,
			&v.ContractSize); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return out, nil
}
