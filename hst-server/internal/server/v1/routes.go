package v1

import "hstserver/model"

// RegisterV1 will register all the v1 routes
func (s *HttpServer) RegisterV1() {

	// Root group with the requests logger and the header reader
	root := s.App.Group("/", s.Middleware.RequestsLogger, s.Middleware.HeaderReader)

	// ------------------------- Auth, unauthenticated -------------------------
	oauth := root.Group("/auth/v1/oauth2")
	oauth.Post("/login", s.Login)
	oauth.Post("/refresh", s.RefreshToken)

	// api group
	api := root.Group("/api")

	// api/v1 group
	v1 := api.Group("/v1")

	// system group
	system := v1.Group("/system")

	// ------------------------- Health -------------------------
	system.Get("/monitor/health", s.CheckSystemHealth)
	system.Get("/monitor/live", s.CheckSystemLive)
	system.Get("/monitor/cache", s.Middleware.Protect, s.Middleware.RequireManager, s.CacheStats)

	// ------------------------- Auth, authenticated -------------------------
	auth := v1.Group("/auth", s.Middleware.Protect)
	auth.Get("/me", s.Me)
	auth.Post("/logout", s.Logout)
	// the one route a restricted session may reach
	auth.Post("/oauth2/change-password", s.ChangePassword)

	// ------------------------- Clients -------------------------
	clients := v1.Group("/clients", s.Middleware.Protect, s.Middleware.RequireManager)
	clients.Get("/", s.Middleware.Authorization(model.MgrRightClientsAccess), s.ListClients)
	clients.Post("/", s.Middleware.Authorization(model.MgrRightClientsCreate), s.CreateClient)
	clients.Get("/:id", s.Middleware.Authorization(model.MgrRightClientsAccess), s.GetClient)
	clients.Patch("/:id", s.Middleware.Authorization(model.MgrRightClientsEdit), s.UpdateClient)
	clients.Delete("/:id", s.Middleware.Authorization(model.MgrRightClientsDelete), s.DeleteClient)

	// ------------------------- Users -------------------------
	users := v1.Group("/users", s.Middleware.Protect, s.Middleware.RequireManager)
	users.Get("/", s.Middleware.Authorization(model.MgrRightAccRead), s.ListUsers)
	users.Post("/", s.Middleware.Authorization(model.MgrRightAccManager), s.CreateUser)
	users.Get("/:login", s.Middleware.Authorization(model.MgrRightAccRead), s.GetUser)
	users.Patch("/:login", s.Middleware.Authorization(model.MgrRightAccManager), s.UpdateUser)
	users.Delete("/:login", s.Middleware.Authorization(model.MgrRightAccDelete), s.DeleteUser)

	// ------------------------- Managers -------------------------
	managers := v1.Group("/managers", s.Middleware.Protect, s.Middleware.RequireManager)
	managers.Get("/", s.Middleware.Authorization(model.MgrRightCfgManagers), s.ListManagers)
	managers.Get("/:login/rights", s.Middleware.Authorization(model.MgrRightCfgManagers), s.GetManagerRights)
}
