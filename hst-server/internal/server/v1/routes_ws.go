package v1

import (
	"context"

	"hstserver/model"
	"hstserver/pkg/cache"
	errs "hstserver/pkg/errors"
	"hstserver/pkg/logger"
	"hstserver/pkg/ws"
)

// RegisterWSV1 maps the socket commands, each to the same right its http route carries.
func (s *HttpServer) RegisterWSV1() {
	s.Hub.RegisterRoute(model.CommandGroupCreate, model.MgrRightCfgGroups, s.CreateGroupWS)
	s.Hub.RegisterRoute(model.CommandGroupUpdate, model.MgrRightCfgGroups, s.UpdateGroupWS)
	s.Hub.RegisterRoute(model.CommandGroupDelete, model.MgrRightCfgGroups, s.DeleteGroupWS)

	s.Log.Log(logger.TypeSys, logger.CodeOK, "websocket commands registered", "commands", s.Hub.Routes())
}

// SessionOf is the live session behind a socket command.
//
// It is read fresh rather than taken from the connection, so a command runs against the access the
// manager holds now and not the access it held when it connected.
func (s *HttpServer) SessionOf(c *ws.Ctx) (*cache.Snapshot, error) {
	snap, err := s.OAuth2.LoadSnapshot(context.Background(), c.Client.SessionId)
	if err != nil {
		return nil, errs.ErrInvalidSession
	}
	return snap, nil
}
