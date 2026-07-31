package server

import (
	"hstserver/internal/server/v1/admin"
	"hstserver/internal/server/v1/trader"
)

// RegisterRoutes will register all the versions routes
func (s *Server) RegisterRoutes() {
	// the shared surface first: it owns the groups the two panels mount under
	root, api := s.Web.RegisterV1()

	admin.New(s.Web).RegisterAdminV1(api)
	trader.New(s.Web).RegisterTraderV1(api, root)
}
