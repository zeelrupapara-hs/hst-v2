package v1

// RegisterV1 will register all the v1 routes
func (s *HttpServer) RegisterV1() {

	// Root group with the requests logger
	root := s.App.Group("/", s.Middleware.RequestsLogger)

	// api group
	api := root.Group("/api")

	// api/v1 group
	v1 := api.Group("/v1")

	// system group
	system := v1.Group("/system")

	// ------------------------- Health -------------------------
	system.Get("/monitor/health", s.CheckSystemHealth)
	system.Get("/monitor/live", s.CheckSystemLive)
}
