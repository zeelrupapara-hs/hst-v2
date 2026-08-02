package trader

import (
	"context"
	"errors"
	"strings"
	"time"

	v1 "hstserver/internal/server/v1"

	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewWatchlist is one Market Watch list with the instruments it draws.
type ViewWatchlist struct {
	WatchlistId int64        `json:"watchlist_id"`
	Login       int64        `json:"login"`
	Name        string       `json:"name"`
	Kind        int32        `json:"kind"`
	SortOrder   int32        `json:"sort_order"`
	CreatedAt   int64        `json:"created_at"`
	UpdatedAt   int64        `json:"updated_at"`
	Symbols     []ViewSymbol `json:"symbols"`
}

type watchlistBody struct {
	Name      *string `json:"name"`
	Kind      *int32  `json:"kind"`
	SortOrder *int32  `json:"sort_order"`
}

type watchlistSymbolsBody struct {
	SymbolIds []int64 `json:"symbol_ids"`
}

var errSymbolNotGranted = errors.New("symbol is not granted to this account's group")

// GetMyWatchlists lists the caller's Market Watch lists, each with its symbols.
//
//	@Id			GetMyWatchlists
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewWatchlist}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/watchlists [get]
func (s *Server) GetMyWatchlists(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readWatchlists(c.UserContext(), snap.Login, 0)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// CreateMyWatchlist adds one empty list to the caller.
//
//	@Id			CreateMyWatchlist
//	@Tags		Trader
//	@Produce	json
//	@Param		body	body		watchlistBody	true	"list"
//	@Success	200		{object}	Response{data=ViewWatchlist}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/watchlists [post]
func (s *Server) CreateMyWatchlist(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body watchlistBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	name := strings.TrimSpace(v1.PtrOr(body.Name, ""))
	if name == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	now := time.Now().UnixNano()
	var id int64
	err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.watchlists (login, name, kind, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5) RETURNING watchlist_id`,
		snap.Login, name, v1.PtrOr(body.Kind, 0), v1.PtrOr(body.SortOrder, 0), now).Scan(&id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.oneWatchlist(c, snap.Login, id)
}

// UpdateMyWatchlist renames or reorders one of the caller's lists.
//
//	@Id			UpdateMyWatchlist
//	@Tags		Trader
//	@Produce	json
//	@Param		watchlist_id	path		int				true	"list id"
//	@Param		body			body		watchlistBody	true	"fields to change"
//	@Success	200				{object}	Response{data=ViewWatchlist}
//	@Failure	400				{object}	Response
//	@Failure	403				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/watchlists/{watchlist_id} [put]
func (s *Server) UpdateMyWatchlist(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := c.ParamsInt("watchlist_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var body watchlistBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`UPDATE hst.watchlists
		    SET name       = COALESCE($3, name),
		        kind       = COALESCE($4, kind),
		        sort_order = COALESCE($5, sort_order),
		        updated_at = $6
		  WHERE watchlist_id = $1 AND login = $2`,
		id, snap.Login, body.Name, body.Kind, body.SortOrder, time.Now().UnixNano())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.oneWatchlist(c, snap.Login, int64(id))
}

// DeleteMyWatchlist removes one of the caller's lists and its symbols.
//
//	@Id			DeleteMyWatchlist
//	@Tags		Trader
//	@Produce	json
//	@Param		watchlist_id	path		int	true	"list id"
//	@Success	200				{object}	Response
//	@Failure	400				{object}	Response
//	@Failure	403				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/watchlists/{watchlist_id} [delete]
func (s *Server) DeleteMyWatchlist(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := c.ParamsInt("watchlist_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.watchlists WHERE watchlist_id = $1 AND login = $2`, id, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, nil)
}

// SetMyWatchlistSymbols replaces the whole symbol set of one list, in the order given.
//
//	@Id			SetMyWatchlistSymbols
//	@Tags		Trader
//	@Produce	json
//	@Param		watchlist_id	path		int						true	"list id"
//	@Param		body			body		watchlistSymbolsBody	true	"symbol ids"
//	@Success	200				{object}	Response{data=ViewWatchlist}
//	@Failure	400				{object}	Response
//	@Failure	403				{object}	Response
//	@Failure	404				{object}	Response
//	@Failure	500				{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/watchlists/{watchlist_id}/symbols [put]
func (s *Server) SetMyWatchlistSymbols(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := c.ParamsInt("watchlist_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var body watchlistSymbolsBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	// a symbol the caller's group was never granted must not enter the list
	var granted int
	err = s.DB.DB.QueryRow(c.UserContext(),
		`SELECT count(*) FROM hst.symbols s
		  WHERE s.symbol_id = ANY($2)
		    AND EXISTS (
		        SELECT 1
		          FROM hst.groups_symbols gs
		          JOIN hst.groups g ON g.group_id = gs.group_id
		          JOIN hst.users u ON u."group" = g."group"
		         WHERE u.login = $1
		           AND (gs.path = '*' OR s.path = gs.path OR starts_with(s.path, rtrim(gs.path, '*')))
		    )`, snap.Login, body.SymbolIds).Scan(&granted)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if granted != len(uniq(body.SymbolIds)) {
		return s.App.HttpResponseBadRequest(c, errSymbolNotGranted)
	}

	tx, err := s.DB.DB.Begin(c.UserContext())
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer func() { _ = tx.Rollback(c.UserContext()) }()

	// ownership is taken inside the transaction and held, so the list cannot change hands mid-write
	var owned bool
	err = tx.QueryRow(c.UserContext(),
		`SELECT true FROM hst.watchlists WHERE watchlist_id = $1 AND login = $2 FOR UPDATE`,
		id, snap.Login).Scan(&owned)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if _, err := tx.Exec(c.UserContext(),
		`DELETE FROM hst.watchlist_symbols WHERE watchlist_id = $1`, id); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	for i, symbolId := range uniq(body.SymbolIds) {
		if _, err := tx.Exec(c.UserContext(),
			`INSERT INTO hst.watchlist_symbols (watchlist_id, symbol_id, sort_order) VALUES ($1, $2, $3)`,
			id, symbolId, i); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
	if _, err := tx.Exec(c.UserContext(),
		`UPDATE hst.watchlists SET updated_at = $2 WHERE watchlist_id = $1`,
		id, time.Now().UnixNano()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if err := tx.Commit(c.UserContext()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.oneWatchlist(c, snap.Login, int64(id))
}

// oneWatchlist answers with a single list after a write.
func (s *Server) oneWatchlist(c *fiber.Ctx, login, id int64) error {
	out, err := s.readWatchlists(c.UserContext(), login, id)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	return s.App.HttpResponseOK(c, out[0])
}

// readWatchlists reads one login's lists with their symbols; id 0 means every list.
func (s *Server) readWatchlists(ctx context.Context, login, id int64) ([]ViewWatchlist, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT w.watchlist_id, w.login, w.name, w.kind, w.sort_order, w.created_at, w.updated_at
		   FROM hst.watchlists w
		  WHERE w.login = $1 AND ($2 = 0 OR w.watchlist_id = $2)
		  ORDER BY w.sort_order, w.watchlist_id`, login, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewWatchlist{}
	index := map[int64]int{}
	for rows.Next() {
		var v ViewWatchlist
		if err := rows.Scan(&v.WatchlistId, &v.Login, &v.Name, &v.Kind, &v.SortOrder,
			&v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		v.Symbols = []ViewSymbol{}
		index[v.WatchlistId] = len(out)
		out = append(out, v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(out) == 0 {
		return out, nil
	}

	srows, err := s.DB.DB.Query(ctx,
		`SELECT ws.watchlist_id, s.symbol_id, s.symbol, s.path, s.description, s.digits,
		        s.trade_mode, s.calc_mode, s.exec_mode, s.spread, s.contract_size, s.date_modified
		   FROM hst.watchlist_symbols ws
		   JOIN hst.symbols s ON s.symbol_id = ws.symbol_id
		   JOIN hst.watchlists w ON w.watchlist_id = ws.watchlist_id
		  WHERE w.login = $1 AND ($2 = 0 OR w.watchlist_id = $2)
		  ORDER BY ws.watchlist_id, ws.sort_order, s.symbol`, login, id)
	if err != nil {
		return nil, err
	}
	defer srows.Close()

	for srows.Next() {
		var listId int64
		var v ViewSymbol
		if err := srows.Scan(&listId, &v.SymbolId, &v.Symbol, &v.Path, &v.Description, &v.Digits,
			&v.TradeMode, &v.CalcMode, &v.ExecMode, &v.Spread, &v.ContractSize, &v.DateModified); err != nil {
			return nil, err
		}
		if i, ok := index[listId]; ok {
			out[i].Symbols = append(out[i].Symbols, v)
		}
	}
	if srows.Err() != nil {
		return nil, srows.Err()
	}

	return out, nil
}

// uniq keeps the first occurrence of each id, so the caller's order survives.
func uniq(ids []int64) []int64 {
	seen := map[int64]bool{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
