package trader

import (
	"context"
	"fmt"
	"sort"
	"strings"

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
	granted := ""
	for i := range list {
		if strings.EqualFold(list[i].Symbol, name) {
			granted = list[i].Symbol
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

	return s.App.HttpResponseOK(c, out)
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
		        COALESCE(o.spread_diff, s.spread_diff),
		        COALESCE(o.spread_diff_balance, s.spread_diff_balance),
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
		        SELECT gs.trade_mode, gs.exec_mode, gs.spread_diff, gs.spread_diff_balance,
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
			&v.TradeMode, &v.CalcMode, &v.ExecMode, &v.SpreadDiff, &v.SpreadDiffBalance,
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
