package v1

import (
	"errors"

	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// closing a position writes a deal and removes the row, so active=false has no table to read
var errClosedPositions = errors.New("closed positions are not kept, read /deals or /orders?active=false for history")

// GetAllPositions lists every open position the manager's group masks reach.
//
//	@Id			GetAllPositions
//	@Tags		Positions
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewPosition}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions [get]
func (s *HttpServer) GetAllPositions(c *fiber.Ctx) error {
	if !c.QueryBool("active", true) {
		return s.App.HttpResponseBadRequest(c, errClosedPositions)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	out, err := s.readPositions(c.UserContext(), where, args)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyPositions lists the calling account's open positions.
//
//	@Id			GetMyPositions
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewPosition}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions [get]
func (s *HttpServer) GetMyPositions(c *fiber.Ctx) error {
	if !c.QueryBool("active", true) {
		return s.App.HttpResponseBadRequest(c, errClosedPositions)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readPositions(c.UserContext(), "p.login = $1", []any{snap.Login})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.overlayLive(c.UserContext(), snap.Login, out)

	return s.App.HttpResponseOK(c, out)
}

// GetAccountPositions lists one named login's open positions.
//
//	@Id			GetAccountPositions
//	@Tags		Positions
//	@Produce	json
//	@Param		login	path		int	true	"the account"
//	@Success	200		{object}	Response{data=[]ViewPosition}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/accounts/{login} [get]
func (s *HttpServer) GetAccountPositions(c *fiber.Ctx) error {
	if !c.QueryBool("active", true) {
		return s.App.HttpResponseBadRequest(c, errClosedPositions)
	}

	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	out, err := s.readPositions(c.UserContext(), "p.login = $1 AND "+where,
		append([]any{int64(login)}, args...))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetPosition reads one position.
//
//	@Id			GetPosition
//	@Tags		Positions
//	@Produce	json
//	@Param		position_id	path		int	true	"the position"
//	@Success	200			{object}	Response{data=ViewPosition}
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/{position_id} [get]
func (s *HttpServer) GetPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	positionId, err := c.ParamsInt("position_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	out, err := s.readPositions(c.UserContext(), "p.position_id = $1 AND "+where,
		append([]any{int64(positionId)}, args...))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, out[0])
}

// UpdatePosition changes the levels of a named login's position.
//
//	@Id			UpdatePosition
//	@Tags		Positions
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int			true	"the position"
//	@Param		body		body		UptPosition	true	"the new levels"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/{position_id} [put]
func (s *HttpServer) UpdatePosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body UptPosition
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.updatePosition(c.UserContext(), &body, snap.Login)
	return s.accepted(c, res, status, err)
}

// UpdateMyPosition changes the levels of the calling account's position.
//
//	@Id			UpdateMyPosition
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int				true	"the position"
//	@Param		body		body		UptMyPosition	true	"the new levels"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions/{position_id} [put]
func (s *HttpServer) UpdateMyPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if isReadOnlyScope(snap.Scope) {
		return s.App.HttpResponseForbidden(c, errs.ErrReadOnlySession)
	}

	var body UptMyPosition
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	res, status, err := s.updatePosition(c.UserContext(), uptPositionFromMy(&body, snap.Login), 0)
	return s.accepted(c, res, status, err)
}

// ClosePosition closes a named login's position.
//
//	@Id			ClosePosition
//	@Tags		Positions
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int				true	"the position"
//	@Param		body		body		ClosePosition	true	"how much to close"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/{position_id}/close [post]
func (s *HttpServer) ClosePosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body ClosePosition
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.closePosition(c.UserContext(), &body, snap.Login)
	return s.accepted(c, res, status, err)
}

// CloseMyPosition closes the calling account's position.
//
//	@Id			CloseMyPosition
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int				true	"the position"
//	@Param		body		body		CloseMyPosition	true	"how much to close"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions/{position_id}/close [post]
func (s *HttpServer) CloseMyPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if isReadOnlyScope(snap.Scope) {
		return s.App.HttpResponseForbidden(c, errs.ErrReadOnlySession)
	}

	var body CloseMyPosition
	_ = c.BodyParser(&body)
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	res, status, err := s.closePosition(c.UserContext(), closeFromMy(&body, snap.Login), 0)
	return s.accepted(c, res, status, err)
}

// CloseByPosition settles a named login's pair of opposite positions.
//
//	@Id			CloseByPosition
//	@Tags		Positions
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int				true	"the position"
//	@Param		body		body		CloseByPosition	true	"the opposite position"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/{position_id}/close-by [post]
func (s *HttpServer) CloseByPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CloseByPosition
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.closeByPosition(c.UserContext(), &body, snap.Login)
	return s.accepted(c, res, status, err)
}

// CloseByMyPosition settles the calling account's own pair.
//
//	@Id			CloseByMyPosition
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		position_id	path		int					true	"the position"
//	@Param		body		body		CloseByMyPosition	true	"the opposite position"
//	@Success	202			{object}	Response{data=Accepted}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/positions/{position_id}/close-by [post]
func (s *HttpServer) CloseByMyPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if isReadOnlyScope(snap.Scope) {
		return s.App.HttpResponseForbidden(c, errs.ErrReadOnlySession)
	}

	var body CloseByMyPosition
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("position_id"); err == nil {
		body.PositionId = int64(id)
	}

	res, status, err := s.closeByPosition(c.UserContext(), closeByFromMy(&body, snap.Login), 0)
	return s.accepted(c, res, status, err)
}
