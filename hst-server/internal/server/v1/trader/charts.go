package trader

import (
	"encoding/json"
	"errors"
	"time"

	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewChartLayout is one chart slot's saved TradingView state.
type ViewChartLayout struct {
	ChartId   int32           `json:"chart_id"`
	Content   json.RawMessage `json:"content"`
	UpdatedAt int64           `json:"updated_at"`
}

type chartBody struct {
	Content json.RawMessage `json:"content"`
}

// maxChartBlob caps one saved state; a real widget.save() blob runs 10-100KB.
const maxChartBlob = 2 << 20

var errChartContent = errors.New("content must be a json document")

// checkChartContent admits the body of a chart or settings write.
func checkChartContent(content json.RawMessage) error {
	if len(content) == 0 || len(content) > maxChartBlob || !json.Valid(content) {
		return errChartContent
	}
	return nil
}

// GetMyCharts lists the caller's saved chart slots.
//
//	@Id			GetMyCharts
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewChartLayout}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/charts [get]
func (s *Server) GetMyCharts(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT chart_id, content, updated_at FROM hst.chart_layouts
		  WHERE login = $1 ORDER BY chart_id`, snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := []ViewChartLayout{}
	for rows.Next() {
		var v ViewChartLayout
		if err := rows.Scan(&v.ChartId, &v.Content, &v.UpdatedAt); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// SetMyChart saves one chart slot's state, replacing whatever the slot held.
//
//	@Id			SetMyChart
//	@Tags		Trader
//	@Produce	json
//	@Param		chart_id	path		int			true	"chart slot, 1-4"
//	@Param		body		body		chartBody	true	"widget.save() state"
//	@Success	200			{object}	Response
//	@Failure	400			{object}	Response
//	@Failure	403			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/charts/{chart_id} [put]
func (s *Server) SetMyChart(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	chartId, err := c.ParamsInt("chart_id")
	if err != nil || chartId < 1 || chartId > 4 {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	var body chartBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if err := checkChartContent(body.Content); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	if _, err := s.DB.DB.Exec(c.UserContext(),
		`INSERT INTO hst.chart_layouts (login, chart_id, content, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (login, chart_id)
		 DO UPDATE SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at`,
		snap.Login, chartId, body.Content, time.Now().UnixNano()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, nil)
}

// DeleteMyChart resets one chart slot.
//
//	@Id			DeleteMyChart
//	@Tags		Trader
//	@Produce	json
//	@Param		chart_id	path		int	true	"chart slot, 1-4"
//	@Success	200			{object}	Response
//	@Failure	400			{object}	Response
//	@Failure	403			{object}	Response
//	@Failure	404			{object}	Response
//	@Failure	500			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/charts/{chart_id} [delete]
func (s *Server) DeleteMyChart(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	chartId, err := c.ParamsInt("chart_id")
	if err != nil || chartId < 1 || chartId > 4 {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.chart_layouts WHERE login = $1 AND chart_id = $2`, snap.Login, chartId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, nil)
}

// GetMyChartSettings reads the caller's shared chart settings document.
//
//	@Id			GetMyChartSettings
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/charts/settings [get]
func (s *Server) GetMyChartSettings(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var content json.RawMessage
	err := s.DB.DB.QueryRow(c.UserContext(),
		`SELECT content FROM hst.chart_settings WHERE login = $1`, snap.Login).Scan(&content)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseOK(c, json.RawMessage(`{}`))
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, content)
}

// SetMyChartSettings replaces the caller's shared chart settings document.
//
//	@Id			SetMyChartSettings
//	@Tags		Trader
//	@Produce	json
//	@Param		body	body		chartBody	true	"settings document"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Failure	500		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/charts/settings [put]
func (s *Server) SetMyChartSettings(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body chartBody
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}
	if err := checkChartContent(body.Content); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	if _, err := s.DB.DB.Exec(c.UserContext(),
		`INSERT INTO hst.chart_settings (login, content, updated_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (login)
		 DO UPDATE SET content = EXCLUDED.content, updated_at = EXCLUDED.updated_at`,
		snap.Login, body.Content, time.Now().UnixNano()); err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, nil)
}
