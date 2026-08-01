package admin

import (
	"context"
	"errors"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// The Dealers tab of a routing rule. A request the rule sends to the desk is offered to
// everyone named here, and whoever answers first settles it.

// CrtRoutingDealer adds one dealer to a rule.
type CrtRoutingDealer struct {
	Login int64 `json:"login" validate:"required,gt=0"`
}

// ViewRoutingDealer is one row of the tab.
type ViewRoutingDealer struct {
	DealerId  int64  `json:"dealer_id"`
	RoutingId int64  `json:"routing_id"`
	Login     int64  `json:"login"`
	Name      string `json:"name"`
}

const routingDealerColumns = `dealer_id, routing_id, login, name`

// ListRoutingDealers is who the rule hands its requests to.
//
//	@Id			ListRoutingDealers
//	@Tags		Routing
//	@Produce	json
//	@Param		id	path		int	true	"the rule"
//	@Success	200	{object}	Response{data=[]ViewRoutingDealer}
//	@Failure	404	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id}/dealers [get]
func (s *Server) ListRoutingDealers(c *fiber.Ctx) error {
	routingId, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	rows, err := s.DB.DB.Query(c.UserContext(),
		`SELECT `+routingDealerColumns+`
		   FROM hst.routing_dealers WHERE routing_id = $1 ORDER BY login`, routingId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	defer rows.Close()

	out := make([]ViewRoutingDealer, 0, 8)

	for rows.Next() {
		var v ViewRoutingDealer
		if err := rows.Scan(&v.DealerId, &v.RoutingId, &v.Login, &v.Name); err != nil {
			return s.App.HttpResponseInternalServerErrorRequest(c, err)
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, rows.Err())
	}

	return s.App.HttpResponseOK(c, out)
}

// CreateRoutingDealer puts a manager on the rule's desk.
//
//	@Id			CreateRoutingDealer
//	@Tags		Routing
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"the rule"
//	@Param		body	body		CrtRoutingDealer	true	"the dealer"
//	@Success	201		{object}	Response{data=ViewRoutingDealer}
//	@Failure	400		{object}	Response
//	@Failure	404		{object}	Response
//	@Failure	409		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id}/dealers [post]
func (s *Server) CreateRoutingDealer(c *fiber.Ctx) error {
	routingId, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	var body CrtRoutingDealer
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if err := s.Validate.Struct(body); err != nil {
		return s.App.HttpResponseBadRequest(c, utils.ValidatorMessage(err))
	}

	// only a manager holding the dealing right belongs on a desk
	name, err := s.dealerName(c.UserContext(), body.Login)
	if err != nil {
		if errors.Is(err, errs.ErrNotADealer) {
			return s.App.HttpResponseBadRequest(c, err)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var v ViewRoutingDealer

	err = s.DB.DB.QueryRow(c.UserContext(),
		`INSERT INTO hst.routing_dealers (routing_id, login, name)
		 VALUES ($1,$2,$3) RETURNING `+routingDealerColumns,
		routingId, body.Login, name).
		Scan(&v.DealerId, &v.RoutingId, &v.Login, &v.Name)

	if err != nil {
		if utils.IsUniqueViolation(err) {
			return s.App.HttpResponseConflict(c, errs.ErrAlreadyExists)
		}
		if utils.IsForeignKeyViolation(err) {
			return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	s.NotifySystem(model.SubjectSystemRoutingUpdated, v)

	return s.App.HttpResponseCreated(c, v)
}

// DeleteRoutingDealer takes a manager off the rule's desk.
//
//	@Id			DeleteRoutingDealer
//	@Tags		Routing
//	@Produce	json
//	@Param		id		path		int	true	"the rule"
//	@Param		login	path		int	true	"the dealer"
//	@Success	200		{object}	Response
//	@Failure	404		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/routing/{id}/dealers/{login} [delete]
func (s *Server) DeleteRoutingDealer(c *fiber.Ctx) error {
	routingId, err := c.ParamsInt("id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}
	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrRequiredParams)
	}

	tag, err := s.DB.DB.Exec(c.UserContext(),
		`DELETE FROM hst.routing_dealers WHERE routing_id = $1 AND login = $2`, routingId, login)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if tag.RowsAffected() == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	s.NotifySystem(model.SubjectSystemRoutingUpdated,
		ViewRoutingDealer{RoutingId: int64(routingId), Login: int64(login)})

	return s.App.HttpResponseOK(c, nil)
}

// dealerName reads the manager's name, and refuses a login that may not deal.
func (s *Server) dealerName(ctx context.Context, login int64) (string, error) {
	var name string
	var read, deal int32

	err := s.DB.DB.QueryRow(ctx,
		`SELECT name, right_trades_read, right_trades_dealer
		   FROM hst.managers WHERE login = $1`, login).
		Scan(&name, &read, &deal)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", errs.ErrNotADealer
	}
	if err != nil {
		return "", err
	}

	if read == 0 || deal == 0 {
		return "", errs.ErrNotADealer
	}

	return name, nil
}
