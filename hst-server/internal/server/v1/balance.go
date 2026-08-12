package v1

import (
	"context"
	"errors"
	"math"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type CrtBalance struct {
	Login    int64            `json:"login" validate:"required,gt=0"`
	Action   model.DealAction `json:"action" validate:"gte=0"`
	Amount   float64          `json:"amount" validate:"required"`
	Comment  string           `json:"comment" validate:"max=64"`
	ExpertId int64            `json:"expert_id"`
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
		AllowNegative: payload.Action == model.DealAction_correction,
	}

	return s.sendBalance(ctx, payload.Login, &model.BalanceEvent{
		EventType: model.BalanceEvent_apply,
		Data:      req,
	})
}

func (s *HttpServer) sendBalance(ctx context.Context, login int64, e *model.BalanceEvent) (*model.TradeResult, int, error) {
	return s.request(ctx, model.SubjectSystemBalance, e)
}

// sendFix asks the engine to write value as the account's balance or credit; no deal results.
func (s *HttpServer) sendFix(ctx context.Context, login int64, action model.DealAction,
	value float64, dealer int64) (*model.TradeResult, int, error) {
	return s.sendBalance(ctx, login, &model.BalanceEvent{
		EventType: model.BalanceEvent_apply,
		Data: &model.BalanceRequest{
			RequestId: s.newRequestId(), Login: login, Action: action,
			Amount: value, Comment: "balance fix", Dealer: dealer, Fix: true,
		},
	})
}

func (s *HttpServer) journalBalance(c *fiber.Ctx, body *CrtBalance) {
	s.JournalEntry(c, model.JournalType_trade, logger.CodeOK, journal.BalanceMsg(body.Login,
		model.BalanceActionName(body.Action), body.Amount), body)
}

// Every money movement is a deal, so the stored balance must equal the sum of the account's
// deals. A mismatch means a crash mid-write or a hand in the database, and the fix writes the
// recomputed value back through the engine, never as a direct DB write.

// ViewBalanceCheck is one account's stored money beside the value its deals add up to.
type ViewBalanceCheck struct {
	Login        int64   `json:"login"`
	Balance      float64 `json:"balance"`
	ValidBalance float64 `json:"valid_balance"`
	Credit       float64 `json:"credit"`
	ValidCredit  float64 `json:"valid_credit"`
	Ok           bool    `json:"ok"`
}

// a cent of float drift across a sum of deals is not a broken account
const balanceTolerance = 0.005

// creditActions is the AffectsCredit set, spelled once for the queries here.
const creditActions = `3, 6, 20`

// CheckBalances recomputes every in-scope account's balance and credit from its deals.
//
//	@Id			CheckBalances
//	@Tags		Balance
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewBalanceCheck}
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/check [get]
func (s *HttpServer) CheckBalances(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	rows, err := s.DB.DB.Query(c.UserContext(), `
		SELECT u.login, u.balance,
		       COALESCE(SUM(d.profit + d.storage + d.fee) FILTER (WHERE d.action NOT IN (`+creditActions+`)), 0),
		       u.credit,
		       COALESCE(SUM(d.profit) FILTER (WHERE d.action IN (`+creditActions+`)), 0)
		  FROM hst.users u
		  LEFT JOIN hst.deals d ON d.login = u.login
		 WHERE `+where+`
		 GROUP BY u.login, u.balance, u.credit
		 ORDER BY u.login`, args...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewBalanceCheck{}
	for rows.Next() {
		var v ViewBalanceCheck
		if err := rows.Scan(&v.Login, &v.Balance, &v.ValidBalance, &v.Credit, &v.ValidCredit); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		v.Ok = math.Abs(v.Balance-v.ValidBalance) < balanceTolerance &&
			math.Abs(v.Credit-v.ValidCredit) < balanceTolerance
		if !v.Ok {
			s.Log.Log(logger.TypeUser, logger.CodeWarn, "invalid balance",
				"login", v.Login, "balance", v.Balance, "valid", v.ValidBalance)
			s.JournalEntry(c, model.JournalType_accounts, logger.CodeWarn,
				journal.BalanceInvalidMsg(v.Login, v.Balance, v.ValidBalance), v)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// FixBalanceBody names the account whose money is written back to what its deals say.
type FixBalanceBody struct {
	Login int64 `json:"login" validate:"required,gt=0"`
}

// FixBalance writes the deals-derived values back through the engine (MT5 semantics: a deal
// would move both sides of the equation, so a fix records no deal — only the journal).
//
//	@Id			FixBalance
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		FixBalanceBody	true	"the account"
//	@Success	200		{object}	Response{data=ViewBalanceCheck}
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/fix [post]
func (s *HttpServer) FixBalance(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body FixBalanceBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	check, err := s.checkOne(c.UserContext(), body.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if check == nil {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if check.Ok {
		return s.App.HttpResponseOK(c, check)
	}

	if math.Abs(check.ValidBalance-check.Balance) >= balanceTolerance {
		if _, _, err := s.sendFix(c.UserContext(), body.Login, model.DealAction_correction,
			check.ValidBalance, snap.Login); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}
	if math.Abs(check.ValidCredit-check.Credit) >= balanceTolerance {
		if _, _, err := s.sendFix(c.UserContext(), body.Login, model.DealAction_credit,
			check.ValidCredit, snap.Login); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
	}

	s.Log.Log(logger.TypeUser, logger.CodeAtt, "balance fixed",
		"actor", snap.Login, "login", body.Login,
		"from", check.Balance, "to", check.ValidBalance)
	s.JournalEntry(c, model.JournalType_accounts, logger.CodeAtt,
		journal.BalanceFixedMsg(body.Login, check.Balance, check.ValidBalance), check)

	fixed, err := s.checkOne(c.UserContext(), body.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, fixed)
}

// checkOne recomputes one account; nil when the login does not exist.
func (s *HttpServer) checkOne(ctx context.Context, login int64) (*ViewBalanceCheck, error) {
	var v ViewBalanceCheck
	err := s.DB.DB.QueryRow(ctx, `
		SELECT u.login, u.balance,
		       COALESCE(SUM(d.profit + d.storage + d.fee) FILTER (WHERE d.action NOT IN (`+creditActions+`)), 0),
		       u.credit,
		       COALESCE(SUM(d.profit) FILTER (WHERE d.action IN (`+creditActions+`)), 0)
		  FROM hst.users u
		  LEFT JOIN hst.deals d ON d.login = u.login
		 WHERE u.login = $1
		 GROUP BY u.login, u.balance, u.credit`, login).
		Scan(&v.Login, &v.Balance, &v.ValidBalance, &v.Credit, &v.ValidCredit)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	v.Ok = math.Abs(v.Balance-v.ValidBalance) < balanceTolerance &&
		math.Abs(v.Credit-v.ValidCredit) < balanceTolerance

	return &v, nil
}

// BulkBalanceBody is one comment applied over up to 500 money operations.
type BulkBalanceBody struct {
	Operations []BulkBalanceOp `json:"operations" validate:"required,min=1,max=500,dive"`
	Comment    string          `json:"comment" validate:"max=64"`
}

// BulkBalanceOp is one row of the batch.
type BulkBalanceOp struct {
	Login  int64            `json:"login" validate:"required,gt=0"`
	Action model.DealAction `json:"action" validate:"gte=0"`
	Amount float64          `json:"amount" validate:"required"`
}

// BulkBalanceResult says what happened to one row of the batch.
type BulkBalanceResult struct {
	Login   int64  `json:"login"`
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// BulkBalance applies a batch of money operations; a refused row does not stop the rest.
//
//	@Id			BulkBalance
//	@Tags		Balance
//	@Accept		json
//	@Produce	json
//	@Param		body	body		BulkBalanceBody	true	"the batch"
//	@Success	200		{object}	Response{data=[]BulkBalanceResult}
//	@Failure	400		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/balance/bulk [post]
func (s *HttpServer) BulkBalance(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body BulkBalanceBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	logins := make([]int64, 0, len(body.Operations))
	for _, op := range body.Operations {
		logins = append(logins, op.Login)
	}

	// one query answers reach for the whole batch instead of a round-trip per row
	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)
	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT u.login FROM hst.users u WHERE u.login = ANY($1) AND `+where,
		append([]any{logins}, args...)...)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	reach := map[int64]bool{}
	for rows.Next() {
		var login int64
		if err := rows.Scan(&login); err != nil {
			rows.Close()
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		reach[login] = true
	}
	rows.Close()
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	s.JournalEntry(c, model.JournalType_trade, logger.CodeOK,
		journal.BulkBalanceQueuedMsg(len(body.Operations), body.Comment), body)

	out := make([]BulkBalanceResult, 0, len(body.Operations))
	done := 0
	for _, op := range body.Operations {
		row := BulkBalanceResult{Login: op.Login}
		if !reach[op.Login] {
			row.Message = "out of scope"
		} else if _, _, err := s.makeBalance(c.UserContext(), &CrtBalance{
			Login:   op.Login,
			Action:  op.Action,
			Amount:  op.Amount,
			Comment: body.Comment,
		}, snap.Login); err != nil {
			row.Message = err.Error()
		} else {
			row.Ok = true
			done++
		}
		out = append(out, row)
	}

	refused := len(out) - done
	code := logger.CodeOK
	if refused > 0 {
		code = logger.CodeWarn
	}
	s.Log.Log(logger.TypeTrade, code, "bulk balance finished",
		"actor", snap.Login, "done", done, "refused", refused)
	s.JournalEntry(c, model.JournalType_trade, code,
		journal.BulkBalanceDoneMsg(done, refused), out)

	return s.App.HttpResponseOK(c, out)
}
