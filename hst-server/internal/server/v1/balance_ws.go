package v1

import (
	"context"

	"hstserver/model"
	nethttp "hstserver/pkg/http"
	"hstserver/pkg/ws"
)

func (s *HttpServer) CreateBalanceWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightAccountant) {
		return nil
	}

	payload := &CrtBalance{}
	if err := c.BodyParser(payload); err != nil {
		return s.wsFail(c, nethttp.StatusBadRequest, err)
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	res, status, err := s.makeBalance(context.Background(), payload, c.Login())
	if err != nil {
		return s.wsFail(c, status, err)
	}
	if res.RetCode != 0 {
		// the engine refused: a bad_request frame carrying the result and its retcode, not an ok
		return c.SendEvent(s.App.WSResponseOK(model.EventBadRequest, res))
	}

	return c.SendEvent(s.App.WSResponseOK(c.Type, res))
}
