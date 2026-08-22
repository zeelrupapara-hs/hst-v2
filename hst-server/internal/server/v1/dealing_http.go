package v1

import (
	"strconv"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// ConfirmRequest fills a queued request at the price the client asked, or at market.
//
//	@Id			ConfirmRequest
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string			true	"the queued request"
//	@Param		body		body		ConfirmRequest	true	"the fill"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/confirm [post]
func (s *HttpServer) ConfirmRequest(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body ConfirmRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.confirmRequest(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// RequoteRequest offers the client a new price.
//
//	@Id			RequoteRequest
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string			true	"the queued request"
//	@Param		body		body		RequoteRequest	true	"the new price"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/requote [post]
func (s *HttpServer) RequoteRequest(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body RequoteRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.requoteRequest(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// RejectRequest refuses with a reason the client sees.
//
//	@Id			RejectRequest
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string			true	"the queued request"
//	@Param		body		body		RejectRequest	true	"why"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/reject [post]
func (s *HttpServer) RejectRequest(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body RejectRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.rejectRequest(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// CancelRequest removes the pending order the request was for.
//
//	@Id			CancelRequest
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string			true	"the queued request"
//	@Param		body		body		CancelRequest	true	"why"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/cancel [post]
func (s *HttpServer) CancelRequest(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CancelRequest
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.cancelRequest(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// ListDealingRequests is the desk's queue: every request waiting for a dealer, within the
// caller's group masks.
//
//	@Id			ListDealingRequests
//	@Tags		Dealing
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewOrder}
//	@Failure	403	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing [get]
func (s *HttpServer) ListDealingRequests(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	where = `o.state = ANY($` + strconv.Itoa(len(args)+1) + `) AND ` + where
	args = append(args, []int32{
		int32(model.OrderState_request_add),
		int32(model.OrderState_request_modify),
		int32(model.OrderState_request_cancel),
	})

	// a dealer who has connected works their own queue; anyone else watching the desk, which
	// is what the supervisor right is for, sees all of it
	online, err := s.dealerOnline(c.UserContext(), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	if online {
		args = append(args, snap.Login)
		where += ` AND EXISTS (SELECT 1 FROM hst.routing_dealers rd
		                        WHERE rd.routing_id = o.routing_id AND rd.login = $` +
			strconv.Itoa(len(args)) + `)`
	} else if !snap.ManagerRights.Has(model.MgrRightTradesSupervisor) {
		// off the desk and not a supervisor: nothing to work on
		return s.App.HttpResponseOK(c, []ViewOrder{})
	}

	out, err := s.readOrders(c.UserContext(), where, args, pageOpts{})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// AcceptRequote takes the price a dealer offered back on a named account.
//
//	@Id			AcceptRequote
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string			true	"the requoted request"
//	@Param		body		body		AcceptRequote	true	"the account"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/accept [post]
func (s *HttpServer) AcceptRequote(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body AcceptRequote
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.acceptRequote(c.UserContext(), &body)
	return s.answer(c, res, status, err)
}

// AcceptMyRequote takes the price a dealer offered back, for the account that is signed in.
//
//	@Id			AcceptMyRequote
//	@Tags		Trading
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string	true	"the requoted request"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/requotes/{request_id}/accept [post]
func (s *HttpServer) AcceptMyRequote(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if isReadOnlyScope(snap.Scope) {
		return s.App.HttpResponseForbidden(c, errs.ErrReadOnlySession)
	}

	body := AcceptRequote{RequestId: c.Params("request_id"), Login: snap.Login}

	res, status, err := s.acceptRequote(c.UserContext(), &body)
	return s.answer(c, res, status, err)
}

// ReturnRequest hands a request back so the next dealer on the rule can take it.
//
//	@Id			ReturnRequest
//	@Tags		Dealing
//	@Accept		json
//	@Produce	json
//	@Param		request_id	path		string				true	"the queued request"
//	@Param		body		body		ReturnRequestBody	true	"the account"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/dealing/{request_id}/return [post]
func (s *HttpServer) ReturnRequest(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body ReturnRequestBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	body.RequestId = c.Params("request_id", body.RequestId)

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.returnRequest(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}
