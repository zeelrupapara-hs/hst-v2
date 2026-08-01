package v1

import (
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// CreateBalance applies any money operation to a named account.
//
//	@Id			CreateBalance
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtBalance	true	"the operation"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	503		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance [post]
func (s *HttpServer) CreateBalance(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CrtBalance
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	if reach, err := s.inReach(c, snap, body.Login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	} else if !reach {
		return s.App.HttpResponseForbidden(c, errs.ErrUnauthorizedToAccessResource)
	}

	res, status, err := s.makeBalance(c.UserContext(), &body, snap.Login)
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	s.journalBalance(c, snap.Login, &body)

	return s.answer(c, res, status, err)
}

// CreateDeposit puts money into an account.
//
//	@Id			CreateDeposit
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtDeposit	true	"the deposit"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/deposit [post]
func (s *HttpServer) CreateDeposit(c *fiber.Ctx) error {
	var body CrtDeposit
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	return s.balanceAs(c, &CrtBalance{
		Login:   body.Login,
		Action:  int32(model.DealAction_balance),
		Amount:  body.Amount,
		Comment: body.Comment,
	})
}

// CreateWithdrawal takes money out of an account.
//
//	@Id			CreateWithdrawal
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtWithdrawal	true	"the withdrawal"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/withdrawal [post]
func (s *HttpServer) CreateWithdrawal(c *fiber.Ctx) error {
	var body CrtWithdrawal
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// the amount is given as a positive number and taken out
	return s.balanceAs(c, &CrtBalance{
		Login:   body.Login,
		Action:  int32(model.DealAction_balance),
		Amount:  -body.Amount,
		Comment: body.Comment,
	})
}

// CreateCredit grants or removes credit, which counts toward margin but is not the client's money.
//
//	@Id			CreateCredit
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtCorrection	true	"the credit, negative to remove"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/credit [post]
func (s *HttpServer) CreateCredit(c *fiber.Ctx) error {
	var body CrtCorrection
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	return s.balanceAs(c, &CrtBalance{
		Login:   body.Login,
		Action:  int32(model.DealAction_credit),
		Amount:  body.Amount,
		Comment: body.Comment,
	})
}

// CreateCorrection adjusts a balance by hand, and is the one operation allowed to leave an
// account short.
//
//	@Id			CreateCorrection
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtCorrection	true	"the correction, signed"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/correction [post]
func (s *HttpServer) CreateCorrection(c *fiber.Ctx) error {
	var body CrtCorrection
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	return s.balanceAs(c, &CrtBalance{
		Login:   body.Login,
		Action:  int32(model.DealAction_correction),
		Amount:  body.Amount,
		Comment: body.Comment,
	})
}

func (s *HttpServer) balanceAs(c *fiber.Ctx, body *CrtBalance) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if reach, err := s.inReach(c, snap, body.Login); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	} else if !reach {
		return s.App.HttpResponseForbidden(c, errs.ErrUnauthorizedToAccessResource)
	}

	res, status, err := s.makeBalance(c.UserContext(), body, snap.Login)
	if err != nil {
		return s.App.HttpResponseStatus(c, status, err)
	}

	s.journalBalance(c, snap.Login, body)

	return s.answer(c, res, status, err)
}
