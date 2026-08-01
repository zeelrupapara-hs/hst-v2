package v1

import (
	"context"
	"strconv"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

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
	Price       float64 `json:"price"`
	Profit      float64 `json:"profit"`
	Storage     float64 `json:"storage"`
	Commission  float64 `json:"commission"`
	Fee         float64 `json:"fee"`
	Time        int64   `json:"time"`
	Comment     string  `json:"comment"`
}

const dealColumns = `d.deal_id, d.login, d.order_id, d.position_id, d.symbol, d.action, d.entry,
	d.reason, d.volume, d.price, d.profit, d.storage, d.commission, d.fee, d.time, d.comment`

const dealFrom = ` FROM hst.deals d JOIN hst.users u ON u.login = d.login WHERE `

// readDeals is the one query behind every deal read handler.
func (s *HttpServer) readDeals(ctx context.Context, where string, args []any, limit int) ([]ViewDeal, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+dealColumns+dealFrom+where+` ORDER BY d.deal_id DESC LIMIT `+strconv.Itoa(limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewDeal{}
	for rows.Next() {
		var v ViewDeal
		if err := rows.Scan(&v.DealId, &v.Login, &v.OrderId, &v.PositionId, &v.Symbol,
			&v.Action, &v.Entry, &v.Reason, &v.VolumeUnits, &v.Price, &v.Profit,
			&v.Storage, &v.Commission, &v.Fee, &v.Time, &v.Comment); err != nil {
			return nil, err
		}
		v.Volume = model.VolumeToLots(v.VolumeUnits)
		out = append(out, v)
	}

	return out, rows.Err()
}

// dealLimit is how many lines a read returns, newest first.
func dealLimit(c *fiber.Ctx) int {
	limit := c.QueryInt("limit", 100)
	if limit < 1 || limit > 500 {
		limit = 100
	}
	return limit
}

// GetAllDeals lists every deal the manager's group masks reach.
//
//	@Id			GetAllDeals
//	@Tags		Deals
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
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

	out, err := s.readDeals(c.UserContext(), where, args, dealLimit(c))
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
//	@Success	200		{object}	Response{data=[]ViewDeal}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/deals [get]
func (s *HttpServer) GetMyDeals(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readDeals(c.UserContext(), "d.login = $1", []any{snap.Login}, dealLimit(c))
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

	out, err := s.readDeals(c.UserContext(), "d.login = $1 AND "+where,
		append([]any{int64(login)}, args...), dealLimit(c))
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
		append([]any{int64(dealId)}, args...), 1)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, out[0])
}
