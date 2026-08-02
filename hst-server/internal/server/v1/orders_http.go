package v1

import (
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"
)

// liveStates are the states an order still sits on the book in.
const liveStates = `o.state IN (0, 1, 3, 7, 8, 9)`

const doneStates = `o.state NOT IN (0, 1, 3, 7, 8, 9)`

// activeClause reads the active flag, which defaults to the working set.
func activeClause(c *fiber.Ctx) string {
	if c.QueryBool("active", true) {
		return liveStates
	}
	return doneStates
}

// orderPage leaves the working set whole and pages the history, which only grows.
func orderPage(c *fiber.Ctx) pageOpts {
	if c.QueryBool("active", true) {
		return readPage(c, 0)
	}
	return readPage(c, 100)
}

// GetAllOrders lists every order the manager's group masks reach.
//
//	@Id			GetAllOrders
//	@Tags		Orders
//	@Produce	json
//	@Param		active	query		bool	false	"working orders only, true by default"
//	@Param		limit	query		int		false	"how many, newest first, history only"
//	@Param		page	query		int		false	"which page, zero based"
//	@Param		from	query		int		false	"unix seconds, inclusive"
//	@Param		to		query		int		false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewOrder}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders [get]
func (s *HttpServer) GetAllOrders(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 1)

	out, err := s.readOrders(c.UserContext(), where+" AND "+activeClause(c), args, orderPage(c))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetMyOrders lists the calling account's own orders.
//
//	@Id			GetMyOrders
//	@Tags		Trader
//	@Produce	json
//	@Param		active	query		bool	false	"working orders only, true by default"
//	@Param		limit	query		int		false	"how many, newest first, history only"
//	@Param		page	query		int		false	"which page, zero based"
//	@Param		from	query		int		false	"unix seconds, inclusive"
//	@Param		to		query		int		false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewOrder}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/orders [get]
func (s *HttpServer) GetMyOrders(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	out, err := s.readOrders(c.UserContext(), "o.login = $1 AND "+activeClause(c), []any{snap.Login}, orderPage(c))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetAccountOrders lists one named login's orders.
//
//	@Id			GetAccountOrders
//	@Tags		Orders
//	@Produce	json
//	@Param		login	path		int		true	"the account"
//	@Param		active	query		bool	false	"working orders only, true by default"
//	@Param		limit	query		int		false	"how many, newest first, history only"
//	@Param		page	query		int		false	"which page, zero based"
//	@Param		from	query		int		false	"unix seconds, inclusive"
//	@Param		to		query		int		false	"unix seconds, exclusive"
//	@Success	200		{object}	Response{data=[]ViewOrder}
//	@Failure	403		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders/accounts/{login} [get]
func (s *HttpServer) GetAccountOrders(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	login, err := c.ParamsInt("login")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	out, err := s.readOrders(c.UserContext(),
		"o.login = $1 AND "+where+" AND "+activeClause(c),
		append([]any{int64(login)}, args...), orderPage(c))
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, out)
}

// GetOrder reads one ticket.
//
//	@Id			GetOrder
//	@Tags		Orders
//	@Produce	json
//	@Param		order_id	path		int	true	"the ticket"
//	@Success	200			{object}	Response{data=ViewOrder}
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders/{order_id} [get]
func (s *HttpServer) GetOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	orderId, err := c.ParamsInt("order_id")
	if err != nil {
		return s.App.HttpResponseBadRequest(c, errs.ErrBadRequest)
	}

	where, args := groupWhere(snap.IsManager, snap.ManagerGroups, 2)

	out, err := s.readOrders(c.UserContext(), "o.order_id = $1 AND "+where,
		append([]any{int64(orderId)}, args...), pageOpts{limit: 1})
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}
	if len(out) == 0 {
		return s.App.HttpResponseNotFound(c, errs.ErrNotFound)
	}

	return s.App.HttpResponseOK(c, out[0])
}

// CreateOrder places an order on behalf of a named login.
//
//	@Id			CreateOrder
//	@Tags		Orders
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtOrder	true	"the order"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	503		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders [post]
func (s *HttpServer) CreateOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CrtOrder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.makeOrder(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// CreateMyOrder places an order for the calling account.
//
//	@Id			CreateMyOrder
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CrtMyOrder	true	"the order"
//	@Success	200		{object}	Response{data=TradeResult}
//	@Failure	400		{object}	Response
//	@Failure	503		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/orders [post]
func (s *HttpServer) CreateMyOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if err := writable(snap); err != nil {
		return s.App.HttpResponseForbidden(c, err)
	}

	var body CrtMyOrder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}

	res, status, err := s.makeOrder(c.UserContext(), crtFromMy(&body, snap.Login), 0)
	return s.answer(c, res, status, err)
}

// UpdateOrder modifies a pending order on behalf of a named login.
//
//	@Id			UpdateOrder
//	@Tags		Orders
//	@Accept		json
//	@Produce	json
//	@Param		order_id	path		int			true	"the ticket"
//	@Param		body		body		UptOrder	true	"the new prices"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders/{order_id} [put]
func (s *HttpServer) UpdateOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body UptOrder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("order_id"); err == nil {
		body.OrderId = int64(id)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.updateOrder(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// UpdateMyOrder modifies the calling account's pending order.
//
//	@Id			UpdateMyOrder
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		order_id	path		int			true	"the ticket"
//	@Param		body		body		UptMyOrder	true	"the new prices"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/orders/{order_id} [put]
func (s *HttpServer) UpdateMyOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if err := writable(snap); err != nil {
		return s.App.HttpResponseForbidden(c, err)
	}

	var body UptMyOrder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("order_id"); err == nil {
		body.OrderId = int64(id)
	}

	res, status, err := s.updateOrder(c.UserContext(), uptFromMy(&body, snap.Login), 0)
	return s.answer(c, res, status, err)
}

// CancelOrder removes a pending order on behalf of a named login.
//
//	@Id			CancelOrder
//	@Tags		Orders
//	@Accept		json
//	@Produce	json
//	@Param		order_id	path		int			true	"the ticket"
//	@Param		body		body		CancelOrder	true	"the account"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/orders/{order_id}/cancel [post]
func (s *HttpServer) CancelOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	var body CancelOrder
	if err := c.BodyParser(&body); err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if id, err := c.ParamsInt("order_id"); err == nil {
		body.OrderId = int64(id)
	}

	if ok, err := s.inReach(c, snap, body.Login); !ok {
		return err
	}

	res, status, err := s.cancelOrder(c.UserContext(), &body, snap.Login)
	return s.answer(c, res, status, err)
}

// CancelMyOrder removes the calling account's pending order.
//
//	@Id			CancelMyOrder
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		order_id	path		int	true	"the ticket"
//	@Success	200			{object}	Response{data=TradeResult}
//	@Failure	400			{object}	Response
//	@Failure	404			{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/orders/{order_id}/cancel [post]
func (s *HttpServer) CancelMyOrder(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return s.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}
	if err := writable(snap); err != nil {
		return s.App.HttpResponseForbidden(c, err)
	}

	var body CancelMyOrder
	_ = c.BodyParser(&body)
	if id, err := c.ParamsInt("order_id"); err == nil {
		body.OrderId = int64(id)
	}

	res, status, err := s.cancelOrder(c.UserContext(),
		&CancelOrder{Login: snap.Login, OrderId: body.OrderId, Comment: body.Comment}, 0)
	return s.answer(c, res, status, err)
}
