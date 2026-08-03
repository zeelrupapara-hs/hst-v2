package v1

import (
	"context"

	"hstserver/model"
	errs "hstserver/pkg/errors"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/ws"
)

// wsFail turns a core refusal into the event the socket receives.
func (s *HttpServer) wsFail(c *ws.Ctx, status int, err error) error {
	switch status {
	case nethttp.StatusBadRequest:
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	case nethttp.StatusNotFound:
		return c.SendEvent(s.App.WSResponseNotFound(c.Type, err))
	case nethttp.StatusForbidden:
		return c.SendEvent(s.App.WSResponseForbidden(c.Type, err))
	default:
		return c.SendEvent(s.App.WSResponseInternalServerErrorRequest(c.Type, err))
	}
}

// wsDealer refuses a socket that may not act on somebody else's account.
func (s *HttpServer) wsDealer(c *ws.Ctx, right uint) bool {
	if c.IsManager() && c.Rights().Has(right) {
		return true
	}
	_ = c.SendEvent(s.App.WSResponseForbidden(c.Type, errs.ErrUnauthorizedToAccessResource))
	return false
}

// wsWritable refuses a socket that authenticated with an investor password.
func (s *HttpServer) wsWritable(c *ws.Ctx) bool {
	if !isReadOnlyScope(c.Scope()) {
		return true
	}
	_ = c.SendEvent(s.App.WSResponseForbidden(c.Type, errs.ErrReadOnlySession))
	return false
}

// wsInReach refuses an account outside the socket's group masks.
func (s *HttpServer) wsInReach(c *ws.Ctx, login int64) bool {
	ok, err := s.accountInReach(context.Background(), c.IsManager(), c.Client.Groups(), login)
	if err != nil {
		_ = c.SendEvent(s.App.WSResponseInternalServerErrorRequest(c.Type, err))
		return false
	}
	if !ok {
		_ = c.SendEvent(s.App.WSResponseForbidden(c.Type, errs.ErrAccountOutOfReach))
		return false
	}
	return true
}

// CreateOrderWS places an order on behalf of a named login.
func (s *HttpServer) CreateOrderWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &CrtOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.makeOrder(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CreateMyOrderWS places an order for the calling account.
func (s *HttpServer) CreateMyOrderWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &CrtMyOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if _, status, err := s.makeOrder(context.Background(), crtFromMy(payload, c.Login()), 0); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// UpdateOrderWS modifies a pending order on behalf of a named login.
func (s *HttpServer) UpdateOrderWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &UptOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.updateOrder(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// UpdateMyOrderWS modifies the calling account's pending order.
func (s *HttpServer) UpdateMyOrderWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &UptMyOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if _, status, err := s.updateOrder(context.Background(), uptFromMy(payload, c.Login()), 0); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CancelOrderWS removes a pending order on behalf of a named login.
func (s *HttpServer) CancelOrderWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &CancelOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.cancelOrder(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CancelMyOrderWS removes the calling account's pending order.
func (s *HttpServer) CancelMyOrderWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &CancelMyOrder{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	cancel := &CancelOrder{Login: c.Login(), OrderId: payload.OrderId, Comment: payload.Comment}
	if _, status, err := s.cancelOrder(context.Background(), cancel, 0); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}
