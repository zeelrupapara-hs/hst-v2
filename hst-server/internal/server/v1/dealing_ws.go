package v1

import (
	"context"

	"hstserver/model"
	"hstserver/pkg/ws"
)

// ConfirmRequestWS fills a queued request.
func (s *HttpServer) ConfirmRequestWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesDealer) {
		return nil
	}

	payload := &ConfirmRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.confirmRequest(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// RequoteRequestWS offers the client a new price.
func (s *HttpServer) RequoteRequestWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesDealer) {
		return nil
	}

	payload := &RequoteRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.requoteRequest(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// RejectRequestWS refuses with a reason the client sees.
func (s *HttpServer) RejectRequestWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesDealer) {
		return nil
	}

	payload := &RejectRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.rejectRequest(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CancelRequestWS removes the pending order the request was for.
func (s *HttpServer) CancelRequestWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesDealer) {
		return nil
	}

	payload := &CancelRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.cancelRequest(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}
