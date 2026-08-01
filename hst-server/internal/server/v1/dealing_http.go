package v1

import (
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
