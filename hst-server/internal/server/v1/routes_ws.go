package v1

import (
	"hstserver/model"
	"hstserver/pkg/logger"
)

// RegisterWSV1 binds every inbound socket event to its handler.
func (s *HttpServer) RegisterWSV1() {
	s.Hub.RegisterRoute(model.EventOrderCreate, s.CreateMyOrderWS)
	s.Hub.RegisterRoute(model.EventOrderDealerCreate, s.CreateOrderWS)
	s.Hub.RegisterRoute(model.EventOrderUpdate, s.UpdateMyOrderWS)
	s.Hub.RegisterRoute(model.EventOrderDealerUpdate, s.UpdateOrderWS)
	s.Hub.RegisterRoute(model.EventOrderCancel, s.CancelMyOrderWS)
	s.Hub.RegisterRoute(model.EventOrderDealerCancel, s.CancelOrderWS)

	s.Hub.RegisterRoute(model.EventPositionUpdate, s.UpdateMyPositionWS)
	s.Hub.RegisterRoute(model.EventPositionDealerUpdate, s.UpdatePositionWS)
	s.Hub.RegisterRoute(model.EventPositionClose, s.CloseMyPositionWS)
	s.Hub.RegisterRoute(model.EventPositionDealerClose, s.ClosePositionWS)
	s.Hub.RegisterRoute(model.EventPositionCloseBy, s.CloseByMyPositionWS)

	s.Hub.RegisterRoute(model.EventBalanceCreate, s.CreateBalanceWS)

	s.Hub.RegisterRoute(model.EventDealerConfirm, s.ConfirmRequestWS)
	s.Hub.RegisterRoute(model.EventDealerRequote, s.RequoteRequestWS)
	s.Hub.RegisterRoute(model.EventDealerReject, s.RejectRequestWS)
	s.Hub.RegisterRoute(model.EventDealerCancel, s.CancelRequestWS)

	s.Hub.SetErrorHandler(s.WSErrorHandler)

	s.Log.Log(logger.TypeNet, logger.CodeOK, "websocket routes registered",
		"routes", len(s.Hub.RouterMap))
}

// WSErrorHandler is what a handler's error becomes on the wire.
func (s *HttpServer) WSErrorHandler(err error) *model.Event {
	return s.App.WSResponseInternalServerErrorRequest("", err)
}
