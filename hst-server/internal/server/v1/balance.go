package v1

import (
	"context"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type CrtBalance struct {
	Login    int64   `json:"login" validate:"required,gt=0"`
	Action   int32   `json:"action" validate:"gte=0"`
	Amount   float64 `json:"amount" validate:"required"`
	Comment  string  `json:"comment" validate:"max=64"`
	ExpertId int64   `json:"expert_id"`
}

type CrtDeposit struct {
	Login   int64   `json:"login" validate:"required,gt=0"`
	Amount  float64 `json:"amount" validate:"required,gt=0"`
	Comment string  `json:"comment" validate:"max=64"`
}

type CrtWithdrawal struct {
	Login   int64   `json:"login" validate:"required,gt=0"`
	Amount  float64 `json:"amount" validate:"required,gt=0"`
	Comment string  `json:"comment" validate:"max=64"`
}

type CrtCorrection struct {
	Login   int64   `json:"login" validate:"required,gt=0"`
	Amount  float64 `json:"amount" validate:"required"`
	Comment string  `json:"comment" validate:"max=64"`
}

func (s *HttpServer) makeBalance(ctx context.Context, payload *CrtBalance, dealer int64) (*model.TradeResult, int, error) {
	if err := s.Validate.Struct(payload); err != nil {
		return nil, nethttp.StatusBadRequest, err
	}

	if !model.IsBalanceAction(payload.Action) {
		return nil, nethttp.StatusBadRequest, errs.ErrInvalidBalanceAction
	}

	if payload.Amount == 0 {
		return nil, nethttp.StatusBadRequest, errs.ErrInvalidAmount
	}

	req := &model.BalanceRequest{
		RequestId: s.newRequestId(),
		Login:     payload.Login,
		Action:    payload.Action,
		Amount:    payload.Amount,
		Comment:   payload.Comment,
		Dealer:    dealer,
		ExpertId:  payload.ExpertId,
		// a correction is the one operation allowed to leave the account short
		AllowNegative: model.DealAction(payload.Action) == model.DealAction_correction,
	}

	return s.sendBalance(ctx, payload.Login, &model.BalanceEvent{
		EventType: model.BalanceEvent_apply,
		Data:      req,
	})
}

func (s *HttpServer) sendBalance(ctx context.Context, login int64, e *model.BalanceEvent) (*model.TradeResult, int, error) {
	return s.request(ctx, model.SubjectSystemBalance, e)
}

func (s *HttpServer) journalBalance(c *fiber.Ctx, actor int64, body *CrtBalance) {
	s.JournalEntry(c, logger.CodeOK, journal.BalanceMsg(actor, body.Login,
		model.BalanceActionName(body.Action), body.Amount), body)
}
