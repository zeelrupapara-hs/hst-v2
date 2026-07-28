package server

// RegisterRoutes will register all the versions routes
func (s *Server) RegisterRoutes() {
	// v1 routes
	s.Web.RegisterV1()
}
