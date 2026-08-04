package v1

import (
	"context"

	"hstserver/model"
	"hstserver/pkg/journal"
	"hstserver/pkg/logger"
	"hstserver/pkg/ws"
)

// UpdatePositionWS changes the levels of a named login's position.
func (s *HttpServer) UpdatePositionWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &UptPosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.updatePosition(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// UpdateMyPositionWS changes the levels of the calling account's position.
func (s *HttpServer) UpdateMyPositionWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &UptMyPosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if _, status, err := s.updatePosition(context.Background(),
		uptPositionFromMy(payload, c.Login()), 0); err != nil {
		return s.wsFail(c, status, err)
	}

	s.JournalWS(c, logger.CodeOK, journal.PositionStopsMsg(c.Login(),
		payload.PositionId, payload.PriceSL, payload.PriceTP), payload)

	return nil
}

// ClosePositionWS closes a named login's position.
func (s *HttpServer) ClosePositionWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &ClosePosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.closePosition(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CloseMyPositionWS closes the calling account's position.
func (s *HttpServer) CloseMyPositionWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &CloseMyPosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if _, status, err := s.closePosition(context.Background(),
		closeFromMy(payload, c.Login()), 0); err != nil {
		return s.wsFail(c, status, err)
	}

	s.JournalWS(c, logger.CodeOK, journal.PositionClosedMsg(c.Login(),
		payload.PositionId, payload.Volume), payload)

	return nil
}

// CloseByPositionWS settles a named login's pair of opposite positions.
func (s *HttpServer) CloseByPositionWS(c *ws.Ctx) error {
	if !s.wsDealer(c, model.MgrRightTradesManager) {
		return nil
	}

	payload := &CloseByPosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if !s.wsInReach(c, payload.Login) {
		return nil
	}

	if _, status, err := s.closeByPosition(context.Background(), payload, c.Login()); err != nil {
		return s.wsFail(c, status, err)
	}

	return nil
}

// CloseByMyPositionWS settles the calling account's own pair.
func (s *HttpServer) CloseByMyPositionWS(c *ws.Ctx) error {
	if !s.wsWritable(c) {
		return nil
	}

	payload := &CloseByMyPosition{}
	if err := c.BodyParser(payload); err != nil {
		return c.SendEvent(s.App.WSResponseBadRequest(c.Type, err))
	}

	if _, status, err := s.closeByPosition(context.Background(),
		closeByFromMy(payload, c.Login()), 0); err != nil {
		return s.wsFail(c, status, err)
	}

	s.JournalWS(c, logger.CodeOK, journal.PositionClosedByMsg(c.Login(),
		payload.PositionId, payload.PositionById), payload)

	return nil
}
