package v1

import (
	"context"
	"errors"
	"math"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
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

	s.overlayLiveAll(c.UserContext(), out)

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

	s.overlayLive(c.UserContext(), int64(login), out)

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

// Every open position is the net of its deals, so its stored volume and open price must equal
// what the deal history adds up to. Editing or deleting a deal breaks that, and the fix writes
// the recomputed values back through the engine, never as a direct DB write.

// ViewPositionCheck is one position's stored volume and price beside what its deals say.
type ViewPositionCheck struct {
	PositionId  int64   `json:"position_id"`
	Login       int64   `json:"login"`
	Symbol      string  `json:"symbol"`
	Digits      int32   `json:"digits"`
	Volume      float64 `json:"volume"`
	ValidVolume float64 `json:"valid_volume"`
	PriceOpen   float64 `json:"price_open"`
	ValidPrice  float64 `json:"valid_price"`
	Ok          bool    `json:"ok"`

	volumeExt      int64
	validVolumeExt int64
}

// positionCheckColumns recomputes each position from its trade deals in extended volume units.
// ponytail: entry inout (reversals) is skipped in the sum; rare, flagged for a manual look.
const positionCheckColumns = `
	SELECT p.position_id, p.login, p.symbol, p.digits,
	       GREATEST(p.volume_ext, p.volume * 10000), p.price_open,
	       COALESCE(SUM(CASE WHEN d.entry = 0 THEN GREATEST(d.volume_ext, d.volume * 10000)
	                         WHEN d.entry IN (1, 3) THEN -GREATEST(d.volume_ext, d.volume * 10000)
	                         ELSE 0 END), 0),
	       COALESCE(SUM(d.price * GREATEST(d.volume_ext, d.volume * 10000)) FILTER (WHERE d.entry = 0)
	                / NULLIF(SUM(GREATEST(d.volume_ext, d.volume * 10000)) FILTER (WHERE d.entry = 0), 0), 0)
	  FROM hst.positions p
	  JOIN hst.users u ON u.login = p.login
	  LEFT JOIN hst.deals d ON d.position_id = p.position_id AND d.action IN (0, 1)`

func scanPositionCheck(row pgx.Row) (*ViewPositionCheck, error) {
	var v ViewPositionCheck
	var validPrice float64
	if err := row.Scan(&v.PositionId, &v.Login, &v.Symbol, &v.Digits,
		&v.volumeExt, &v.PriceOpen, &v.validVolumeExt, &validPrice); err != nil {
		return nil, err
	}

	v.Volume = model.ExtToLots(v.volumeExt)
	v.ValidVolume = model.ExtToLots(v.validVolumeExt)
	v.ValidPrice = validPrice
	// half a point of float drift across a weighted average is not a broken position
	v.Ok = v.volumeExt == v.validVolumeExt &&
		math.Abs(v.PriceOpen-v.ValidPrice) < math.Pow(10, -float64(v.Digits))/2

	return &v, nil
}

// CheckPositions recomputes every in-scope open position from its deals.
//
//	@Id			CheckPositions
//	@Tags		Positions
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewPositionCheck}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/check [get]
func (s *HttpServer) CheckPositions(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	rows, err := s.DB.DB.Query(c.UserContext(), positionCheckColumns+`
	 WHERE `+where+`
	 GROUP BY p.position_id, p.login, p.symbol, p.digits, p.volume, p.volume_ext, p.price_open
	 ORDER BY p.position_id`, args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewPositionCheck{}
	for rows.Next() {
		v, err := scanPositionCheck(rows)
		if err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		if !v.Ok {
			s.Log.Log(logger.TypeTrade, logger.CodeWarn, "invalid position",
				"login", v.Login, "position", v.PositionId,
				"volume", v.Volume, "valid", v.ValidVolume)
			s.JournalEntry(c, model.JournalType_trade, logger.CodeWarn,
				journal.PositionInvalidMsg(v.PositionId, v.Volume, v.ValidVolume, v.PriceOpen, v.ValidPrice), v)
		}
		out = append(out, *v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// FixPositionBody names the position whose volume and price are written back from its deals.
type FixPositionBody struct {
	PositionId int64 `json:"position_id" validate:"required,gt=0"`
}

// FixPosition writes the deals-derived volume and open price back through the engine.
// A recomputed volume of zero deletes the position, exactly as MT5 does.
//
//	@Id			FixPosition
//	@Tags		Positions
//	@Accept		json
//	@Produce	json
//	@Param		body	body		FixPositionBody	true	"the position"
//	@Success	200		{object}	Response{data=ViewPositionCheck}
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/fix [post]
func (s *HttpServer) FixPosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body FixPositionBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	check, err := s.checkOnePosition(c.UserContext(), body.PositionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if check == nil {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if ok, err := s.inReach(c, snap, check.Login); !ok {
		return err
	}
	if check.Ok {
		return s.App.HttpResponseOK(c, check)
	}

	if _, _, err := s.request(c.UserContext(), model.SubjectSystemPositions, &model.PositionEvent{
		EventType: model.PositionEvent_fix,
		Data: &model.TradeRequest{
			RequestId: s.newRequestId(), Login: check.Login, PositionId: check.PositionId,
			Volume: check.validVolumeExt / model.ExtPerUnit, Price: check.ValidPrice,
			Comment: "position fix", Dealer: snap.Login,
		},
	}); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeTrade, logger.CodeAtt, "position fixed",
		"actor", snap.Login, "login", check.Login, "position", check.PositionId,
		"volume", check.ValidVolume, "price", check.ValidPrice)
	s.JournalEntry(c, model.JournalType_trade, logger.CodeAtt,
		journal.PositionFixedMsg(check.PositionId, check.Volume, check.ValidVolume), check)

	fixed, err := s.checkOnePosition(c.UserContext(), body.PositionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if fixed == nil {
		// the deals said zero volume, so the engine deleted the position
		check.Volume, check.Ok = 0, true
		return s.App.HttpResponseOK(c, check)
	}

	return s.App.HttpResponseOK(c, fixed)
}

// DeletePosition removes a position outright, without generating an order or a deal.
//
//	@Id			DeletePosition
//	@Tags		Positions
//	@Produce	json
//	@Param		position_id	path		int	true	"the position"
//	@Success	200			{object}	Response{data=Accepted}
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/positions/{position_id} [delete]
func (s *HttpServer) DeletePosition(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	id, err := c.ParamsInt("position_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	check, err := s.checkOnePosition(c.UserContext(), int64(id))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if check == nil {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if ok, err := s.inReach(c, snap, check.Login); !ok {
		return err
	}

	res, _, err := s.request(c.UserContext(), model.SubjectSystemPositions, &model.PositionEvent{
		EventType: model.PositionEvent_delete,
		Data: &model.TradeRequest{
			RequestId: s.newRequestId(), Login: check.Login, PositionId: check.PositionId,
			Comment: "position delete", Dealer: snap.Login,
		},
	})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.Log.Log(logger.TypeTrade, logger.CodeAtt, "position deleted",
		"actor", snap.Login, "login", check.Login, "position", check.PositionId)
	s.JournalEntry(c, model.JournalType_trade, logger.CodeAtt,
		journal.PositionDeletedMsg(check.PositionId, check.Login), check)

	return s.App.HttpResponseOK(c, res)
}

// checkOnePosition recomputes one position; nil when it does not exist.
func (s *HttpServer) checkOnePosition(ctx context.Context, positionId int64) (*ViewPositionCheck, error) {
	v, err := scanPositionCheck(s.DB.DB.QueryRow(ctx, positionCheckColumns+`
	 WHERE p.position_id = $1
	 GROUP BY p.position_id, p.login, p.symbol, p.digits, p.volume, p.volume_ext, p.price_open`,
		positionId))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return v, err
}
