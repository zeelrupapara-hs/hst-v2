package v1

import (
	"context"
	"errors"
	"time"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// historyMonths is how far back each limit_history setting reaches, zero being no limit at all.
var historyMonths = map[model.HistoryLimit]int{
	model.HistoryLimit_months_1: 1,
	model.HistoryLimit_months_3: 3,
	model.HistoryLimit_months_6: 6,
	model.HistoryLimit_year_1:   12,
	model.HistoryLimit_year_2:   24,
	model.HistoryLimit_year_3:   36,
}

// historyFloor is the earliest moment, in unix seconds, this session may read its own history.
//
// Staff read the whole ledger, so only a trading account's group limits it, and an unknown group
// or an unset limit means no floor.
func (s *HttpServer) historyFloor(c *fiber.Ctx) (int64, error) {
	snap, ok := utils.GetClient(c)
	if !ok {
		return 0, errs.ErrCouldNotParseClientCfg
	}
	if snap.IsManager {
		return 0, nil
	}

	var limit model.HistoryLimit
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT limit_history FROM hst.groups WHERE "group" = $1`, snap.Group).Scan(&limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	months, ok := historyMonths[limit]
	if !ok {
		return 0, nil
	}

	return time.Now().AddDate(0, -months, 0).Unix(), nil
}

// historyPage is readPage clamped to the window the caller's group keeps.
func (s *HttpServer) historyPage(c *fiber.Ctx, def int) (pageOpts, error) {
	p := readPage(c, def)

	floor, err := s.historyFloor(c)
	if err != nil {
		return p, err
	}
	if floor > 0 && p.from < floor {
		p.from = floor
	}

	return p, nil
}

// ViewDeal is one line of the ledger.
type ViewDeal struct {
	DealId      int64   `json:"deal_id"`
	Login       int64   `json:"login"`
	OrderId     int64   `json:"order_id"`
	PositionId  int64   `json:"position_id"`
	Symbol      string  `json:"symbol"`
	Action      int32   `json:"action"`
	Entry       int32   `json:"entry"`
	Reason      int32   `json:"reason"`
	Volume      float64 `json:"volume"`
	VolumeUnits int64   `json:"-"`
	VolumeExt   int64   `json:"-"`
	Price       float64 `json:"price"`
	Profit      float64 `json:"profit"`
	Storage     float64 `json:"storage"`
	Commission  float64 `json:"commission"`
	Fee         float64 `json:"fee"`
	Time        int64   `json:"time"`
	Comment     string  `json:"comment"`
}

const dealColumns = `d.deal_id, d.login, d.order_id, d.position_id, d.symbol, d.action, d.entry,
	d.reason, d.volume, d.volume_ext, d.price, d.profit, d.storage, d.commission, d.fee, d.time, d.comment`

const dealFrom = ` FROM hst.deals d JOIN hst.users u ON u.login = d.login WHERE `

// readDeals is the one query behind every deal read handler.
func (s *HttpServer) readDeals(ctx context.Context, where string, args []any, p pageOpts) ([]ViewDeal, error) {
	where, args = p.bound("d.time", where, args)

	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+dealColumns+dealFrom+where+p.tail("d.deal_id"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewDeal{}
	for rows.Next() {
		var v ViewDeal
		if err := rows.Scan(&v.DealId, &v.Login, &v.OrderId, &v.PositionId, &v.Symbol,
			&v.Action, &v.Entry, &v.Reason, &v.VolumeUnits, &v.VolumeExt, &v.Price, &v.Profit,
			&v.Storage, &v.Commission, &v.Fee, &v.Time, &v.Comment); err != nil {
			return nil, err
		}
		v.Volume = model.ExtToLots(model.ExtendedVolume(v.VolumeUnits, v.VolumeExt))
		out = append(out, v)
	}

	return out, rows.Err()
}

// GetAllDeals lists every deal the manager's group masks reach.
//
//	@Id			GetAllDeals
//	@Tags		Deals
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
//	@Param		page	query		int	false	"which page, zero based"
//	@Param		from	query		int	false	"unix seconds, inclusive"
//	@Param		to		query		int	false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewDeal}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/deals [get]
func (s *HttpServer) GetAllDeals(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	out, err := s.readDeals(c.UserContext(), where, args, readPage(c, 100))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyDeals lists the calling account's history.
//
//	@Id			GetMyDeals
//	@Tags		Trader
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
//	@Param		page	query		int	false	"which page, zero based"
//	@Param		from	query		int	false	"unix seconds, inclusive"
//	@Param		to		query		int	false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewDeal}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/deals [get]
func (s *HttpServer) GetMyDeals(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	p, err := s.historyPage(c, 100)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	out, err := s.readDeals(c.UserContext(), "d.login = $1", []any{snap.Login}, p)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetAccountDeals lists one named login's history.
//
//	@Id			GetAccountDeals
//	@Tags		Deals
//	@Produce	json
//	@Param		login	path		int	true	"the account"
//	@Param		limit	query		int	false	"how many, newest first"
//	@Param		page	query		int	false	"which page, zero based"
//	@Param		from	query		int	false	"unix seconds, inclusive"
//	@Param		to		query		int	false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewDeal}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/deals/accounts/{login} [get]
func (s *HttpServer) GetAccountDeals(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	p, err := s.historyPage(c, 100)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	out, err := s.readDeals(c.UserContext(), "d.login = $1 AND "+where,
		append([]any{int64(login)}, args...), p)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetDeal reads one line of the ledger.
//
//	@Id			GetDeal
//	@Tags		Deals
//	@Produce	json
//	@Param		deal_id	path		int	true	"the deal"
//	@Success	200		{object}	Response{data=ViewDeal}
//	@Failure	404		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/deals/{deal_id} [get]
func (s *HttpServer) GetDeal(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	dealId, err := c.ParamsInt("deal_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	out, err := s.readDeals(c.UserContext(), "d.deal_id = $1 AND "+where,
		append([]any{int64(dealId)}, args...), pageOpts{limit: 1})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, out[0])
}
