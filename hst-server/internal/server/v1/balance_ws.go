package v1

import (
	"context"

	"hstserver/model"
	"hstserver/pkg/ws"
)

func (s *HttpServer) CreateBalanceWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightAccountant) {
		return nil
	}

	payload := &CrtBalance{}
	if err := c.BodyParser(payload); err != nil {
		return s.wsFail(c, 400, err)
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	res, status, err := s.makeBalance(context.Background(), payload, c.Login())
	if err != nil {
		return s.wsFail(c, status, err)
	}

	return c.SendEvent(s.App.WSResponseOK(c.Type, res))
}
