package trader

import (
	"context"
	"errors"
	"time"

	v1 "hstserver/internal/server/v1"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ViewAlert is one alert as its owner sees it.
type ViewAlert struct {
	AlertId     int64   `json:"alert_id"`
	Login       int64   `json:"login"`
	Symbol      string  `json:"symbol,omitempty"`
	Kind        int32   `json:"kind"`
	Condition   int32   `json:"condition"`
	Value       float64 `json:"value"`
	Enabled     bool    `json:"enabled"`
	TriggeredAt int64   `json:"triggered_at"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
	Comment     string  `json:"comment"`
}

// CrtAlert is what a terminal sends to arm an alert, and to rewrite one.
type CrtAlert struct {
	Symbol    string  `json:"symbol"`
	Kind      int32   `json:"kind"`
	Condition int32   `json:"condition"`
	Value     float64 `json:"value"`
	Enabled   *bool   `json:"enabled"`
	Comment   string  `json:"comment"`
}

const alertColumns = `alert_id, login, COALESCE(symbol, ''), kind, condition, value, enabled,
	triggered_at, created_at, updated_at, comment`

// readAlerts is the one query behind every alert read handler.
func (s *Server) readAlerts(ctx context.Context, where string, args []any) ([]ViewAlert, error) {
	rows, err := s.DB.DB.Query(ctx,
		`SELECT `+alertColumns+` FROM hst.alerts WHERE `+where+` ORDER BY alert_id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ViewAlert{}
	for rows.Next() {
		var v ViewAlert
		if err := rows.Scan(&v.AlertId, &v.Login, &v.Symbol, &v.Kind, &v.Condition, &v.Value,
			&v.Enabled, &v.TriggeredAt, &v.CreatedAt, &v.UpdatedAt, &v.Comment); err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, rows.Err()
}

// checkAlert is what an alert must be for this login before it is written.
func (s *Server) checkAlert(ctx context.Context, login int64, p *CrtAlert) error {
	if !v1.AlertKindValid(p.Kind) || !v1.AlertCondValid(p.Condition) {
		return errs.ErrBadRequest
	}

	if !v1.AlertIsPrice(p.Kind) {
		// an account alert watches money, which belongs to no instrument
		if p.Symbol != "" {
			return errs.ErrBadRequest
		}
		return nil
	}

	if p.Symbol == "" {
		return errs.ErrBadRequest
	}

	// an instrument the account's group was never granted is not one it may be alerted on
	var granted bool
	err := s.DB.DB.QueryRow(ctx,
		`SELECT EXISTS (
		        SELECT 1
		          FROM hst.symbols s
		          JOIN hst.groups_symbols gs ON (gs.path = '*' OR s.path = gs.path OR starts_with(s.path, rtrim(gs.path, '*')))
		          JOIN hst.groups g ON g.group_id = gs.group_id
		          JOIN hst.users u ON u."group" = g."group"
		         WHERE u.login = $1 AND s.symbol = $2)`, login, p.Symbol).Scan(&granted)
	if err != nil {
		return err
	}
	if !granted {
		return errs.ErrUnauthorizedToAccessResource
	}

	return nil
}

// GetMyAlerts lists the calling account's own alerts.
//
//	@Id			GetMyAlerts
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response{data=[]ViewAlert}
//	@Failure	403	{object}	Response
//	@Failure	500	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/alerts [get]
func (s *Server) GetMyAlerts(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readAlerts(c.UserContext(), "login = $1", []any{snap.Login})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyAlert reads one of the calling account's alerts.
//
//	@Id			GetMyAlert
//	@Tags		Trader
//	@Produce	json
//	@Param		alert_id	path		int	true	"the alert"
//	@Success	200			{object}	Response{data=ViewAlert}
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/alerts/{alert_id} [get]
func (s *Server) GetMyAlert(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	alertId, err := c.ParamsInt("alert_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	out, err := s.readAlerts(c.UserContext(), "alert_id = $1 AND login = $2",
		[]any{int64(alertId), snap.Login})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, out[0])
}

// CreateMyAlert arms a new alert for the calling account.
//
//	@Id			CreateMyAlert
//	@Tags		Trader
//	@Produce	json
//	@Param		payload	body		CrtAlert	true	"the alert"
//	@Success	200		{object}	Response{data=ViewAlert}
//	@Failure	400		{object}	Response
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/alerts [post]
func (s *Server) CreateMyAlert(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	p := &CrtAlert{}
	if err := c.BodyParser(p); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	if err := s.checkAlert(c.UserContext(), snap.Login, p); err != nil {
		return s.alertCheckFailed(c, err)
	}

	enabled := p.Enabled == nil || *p.Enabled
	now := time.Now().UnixNano()

	v := ViewAlert{}
	err := s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.alerts (login, symbol, kind, condition, value, enabled, created_at, updated_at, comment)
		 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $7, $8)
		 RETURNING `+alertColumns,
		snap.Login, p.Symbol, p.Kind, p.Condition, p.Value, enabled, now, p.Comment).
		Scan(&v.AlertId, &v.Login, &v.Symbol, &v.Kind, &v.Condition, &v.Value,
			&v.Enabled, &v.TriggeredAt, &v.CreatedAt, &v.UpdatedAt, &v.Comment)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.reloadAlerts(c)

	return s.App.HttpResponseOK(c, v)
}

// UpdateMyAlert rewrites one of the calling account's alerts.
//
//	@Id			UpdateMyAlert
//	@Tags		Trader
//	@Produce	json
//	@Param		alert_id	path		int			true	"the alert"
//	@Param		payload		body		CrtAlert	true	"the alert"
//	@Success	200			{object}	Response{data=ViewAlert}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/alerts/{alert_id} [put]
func (s *Server) UpdateMyAlert(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	alertId, err := c.ParamsInt("alert_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	p := &CrtAlert{}
	if err := c.BodyParser(p); err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	if err := s.checkAlert(c.UserContext(), snap.Login, p); err != nil {
		return s.alertCheckFailed(c, err)
	}

	enabled := p.Enabled == nil || *p.Enabled

	v := ViewAlert{}
	err = s.DB.DB.QueryRow(c.UserContext(),
		`UPDATE hst.alerts
		    SET symbol = NULLIF($3, ''), kind = $4, condition = $5, value = $6, enabled = $7,
		        updated_at = $8, comment = $9
		  WHERE alert_id = $1 AND login = $2
		 RETURNING `+alertColumns,
		int64(alertId), snap.Login, p.Symbol, p.Kind, p.Condition, p.Value, enabled,
		time.Now().UnixNano(), p.Comment).
		Scan(&v.AlertId, &v.Login, &v.Symbol, &v.Kind, &v.Condition, &v.Value,
			&v.Enabled, &v.TriggeredAt, &v.CreatedAt, &v.UpdatedAt, &v.Comment)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.reloadAlerts(c)

	return s.App.HttpResponseOK(c, v)
}

// DeleteMyAlert drops one of the calling account's alerts.
//
//	@Id			DeleteMyAlert
//	@Tags		Trader
//	@Produce	json
//	@Param		alert_id	path		int	true	"the alert"
//	@Success	200			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/alerts/{alert_id} [delete]
func (s *Server) DeleteMyAlert(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	alertId, err := c.ParamsInt("alert_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.alerts WHERE alert_id = $1 AND login = $2`, int64(alertId), snap.Login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	s.reloadAlerts(c)

	return s.App.HttpResponseOK(c, fiber.Map{"alert_id": alertId})
}

// alertCheckFailed is what a rejected alert body answers with.
func (s *Server) alertCheckFailed(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errs.ErrBadRequest):
		return s.App.HttpResponseBadRequest(c, err)
	case errors.Is(err, errs.ErrUnauthorizedToAccessResource):
		return s.App.HttpResponseForbidden(c, err)
	default:
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
}

// reloadAlerts keeps the evaluator current after a write, on this pod now and on the others by event.
func (s *Server) reloadAlerts(c *fiber.Ctx) {
	if err := s.ReloadAlerts(c.UserContext()); err != nil {
		s.Log.Log(logger.TypeSys, logger.CodeErr, "could not reload alerts", "error", err.Error())
	}
	snap, _ := utils.GetClient(c)
	s.NotifySystem(v1.SubjectSystemAlertsUpdated, fiber.Map{"login": snap.Login})
}
