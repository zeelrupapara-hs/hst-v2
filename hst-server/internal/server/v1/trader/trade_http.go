package trader

import (
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	v1 "hstserver/internal/server/v1"

	"github.com/gofiber/fiber/v2"
)

// The trader's trading routes.
//
// A trader may only trade its own account, and the login comes from the session, never from the
// request. There is nowhere in these handlers to name somebody else's account.

// CrtTrade is a trade request, aliased for swagger.
type CrtTrade = v1.CrtTrade

// TradeResult is the engine's answer, aliased for the same reason.
type TradeResult = v1.TradeResult

// MyTrade sends a trade for the calling account.
//
//	@Id				MyTrade
//	@Description	Open, close or place an order on the calling account. Volume is in lots.
//	@Tags			Trader
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CrtTrade	true	"the trade to make"
//	@Success		200		{object}	Response{data=TradeResult}
//	@Failure		400		{object}	Response{data=TradeResult}
//	@Failure		403		{object}	Response
//	@Failure		503		{object}	Response
//	@Security		BearerAuth
//	@Router			/api/trader/v1/trade [post]
func (s *Server) MyTrade(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	// an investor password may look at everything and change nothing
	if snap.Scope == int32(model.UsersPasswords_investor) {
		return s.App.HttpResponseForbidden(c, errs.ErrReadOnlySession)
	}

	var body CrtTrade
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// the reason records how the trade arrived, which the routing rules can key on
	res, status, err := s.SendTrade(c.UserContext(), snap.Login, &body,
		model.OrderReason_client, 0)
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	// a refusal by the engine is still an answer: it carries the reason the client needs
	if res.RetCode != 0 {
		return s.App.HttpResponseStatus(c, status, errs.New(res.Message))
	}

	return s.App.HttpResponseOK(c, res)
}

// MyOrders lists the calling account's working orders.
//
//	@Id			MyOrders
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewOrder}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/orders [get]
func (s *Server) MyOrders(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT order_id, symbol, type, state, volume_initial, volume_current,
		        price_order, price_current, price_sl, price_tp,
		        time_setup, time_expiration, comment
		   FROM hst.orders
		  WHERE login = $1 AND state IN (0, 1, 3)
		  ORDER BY order_id DESC`, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewOrder{}
	for rows.Next() {
		var v ViewOrder
		if err := rows.Scan(&v.OrderId, &v.Symbol, &v.Type, &v.State,
			&v.VolumeInitial, &v.VolumeCurrent, &v.PriceOrder, &v.PriceCurrent,
			&v.PriceSL, &v.PriceTP, &v.TimeSetup, &v.TimeExpiration, &v.Comment); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Volume = model.VolumeToLots(v.VolumeCurrent)
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// MyPositions lists the calling account's open positions.
//
//	@Id			MyPositions
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewPosition}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions [get]
func (s *Server) MyPositions(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT position_id, symbol, action, volume, price_open, price_current,
		        price_sl, price_tp, profit, storage, time_create, comment
		   FROM hst.positions
		  WHERE login = $1
		  ORDER BY position_id DESC`, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewPosition{}
	for rows.Next() {
		var v ViewPosition
		if err := rows.Scan(&v.PositionId, &v.Symbol, &v.Action, &v.VolumeUnits,
			&v.PriceOpen, &v.PriceCurrent, &v.PriceSL, &v.PriceTP,
			&v.Profit, &v.Storage, &v.TimeCreate, &v.Comment); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Volume = model.VolumeToLots(v.VolumeUnits)
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// MyDeals lists the calling account's history.
//
//	@Id			MyDeals
//	@Tags		Trader
//	@Produce	json
//	@Param		limit	query		int	false	"how many, newest first"
//	@Success	200		{object}	Response{data=[]ViewDeal}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/deals [get]
func (s *Server) MyDeals(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	limit := c.QueryInt("limit", 100)
	if limit < 1 || limit > 500 {
		limit = 100
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT deal_id, order_id, position_id, symbol, action, entry, volume,
		        price, profit, storage, commission, time, comment
		   FROM hst.deals
		  WHERE login = $1
		  ORDER BY deal_id DESC
		  LIMIT $2`, snap.Login, limit)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewDeal{}
	for rows.Next() {
		var v ViewDeal
		if err := rows.Scan(&v.DealId, &v.OrderId, &v.PositionId, &v.Symbol,
			&v.Action, &v.Entry, &v.VolumeUnits, &v.Price, &v.Profit,
			&v.Storage, &v.Commission, &v.Time, &v.Comment); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Volume = model.VolumeToLots(v.VolumeUnits)
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}
