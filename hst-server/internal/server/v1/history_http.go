package v1

import (
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// GetHistory returns chart bars for any instrument, for the back office.
//
//	@Id			GetHistory
//	@Tags		History
//	@Produce	json
//	@Param		symbol		query		string	true	"the instrument"
//	@Param		resolution	query		string	true	"1, 5, 15, 30, 60, 240, D, W or M"
//	@Param		from		query		int		true	"start, epoch seconds"
//	@Param		to			query		int		true	"end, epoch seconds"
//	@Param		countback	query		int		false	"only the newest this many bars"
//	@Success	200			{object}	Bars
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	503			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/history [get]
func (s *HttpServer) GetHistory(c *fiber.Ctx) error {
	q, err := s.chartQuery(c)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	symbolId, err := s.symbolByName(c.UserContext(), q.Symbol)
	if err != nil {
		return s.chartFailed(c, err)
	}

	return s.answerHistory(c, symbolId, q)
}

// GetMyHistory returns chart bars for an instrument the signed-in account may trade.
//
//	@Id			GetMyHistory
//	@Tags		Trading
//	@Produce	json
//	@Param		symbol		query		string	true	"the instrument"
//	@Param		resolution	query		string	true	"1, 5, 15, 30, 60, 240, D, W or M"
//	@Param		from		query		int		true	"start, epoch seconds"
//	@Param		to			query		int		true	"end, epoch seconds"
//	@Param		countback	query		int		false	"only the newest this many bars"
//	@Success	200			{object}	Bars
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	503			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/history [get]
func (s *HttpServer) GetMyHistory(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	q, err := s.chartQuery(c)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	// an instrument the account's group was never granted is not one it can chart
	symbolId, err := s.symbolForTrader(c.UserContext(), snap.Login, q.Symbol)
	if err != nil {
		return s.chartFailed(c, err)
	}

	return s.answerHistory(c, symbolId, q)
}

func (s *HttpServer) chartQuery(c *fiber.Ctx) (*HistoryQuery, error) {
	return readHistoryQuery(
		c.Query("symbol"),
		c.Query("resolution"),
		c.Query("from"),
		c.Query("to"),
		c.Query("countback"),
	)
}

func (s *HttpServer) answerHistory(c *fiber.Ctx, symbolId int64, q *HistoryQuery) error {
	bars, err := s.history(c.UserContext(), symbolId, q)
	if err != nil {
		return s.chartFailed(c, err)
	}

	// the charting contract carries its own status, so the envelope is not used here
	return c.JSON(bars)
}

func (s *HttpServer) chartFailed(c *fiber.Ctx, err error) error {
	if historyUnavailable(err) {
		return s.App.HttpResponseStatus(c, fiber.StatusServiceUnavailable, errs.ErrHistoryUnavailable)
	}
	if err == errs.ErrNotFound {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseInternalServerErrorRequest(c, err)
}
