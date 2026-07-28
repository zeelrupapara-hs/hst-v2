package v1

import (
	"hstserver/model"

	"github.com/gofiber/swagger"
)

// RegisterV1 will register all the v1 routes
func (s *HttpServer) RegisterV1() {

	// Root group with the requests logger and the header reader
	root := s.App.Group("/", s.Middleware.RequestsLogger, s.Middleware.HeaderReader)

	oauth := root.Group("/auth/v1/oauth2")
	oauth.Post("/login", s.Middleware.BasicAuthParser, s.Login)
	oauth.Post("/refresh", s.RefreshToken)

	// swagger ui, served from the generated swagger package
	root.Get("/swagger/*", swagger.HandlerDefault)

	// api group
	api := root.Group("/api")

	// api/v1 group
	v1 := api.Group("/v1")

	// system group
	system := v1.Group("/system")

	// health-check
	system.Get("/monitor/health", s.CheckSystemHealth)
	system.Get("/monitor/live", s.CheckSystemLive)
	system.Get("/monitor/cache", s.Middleware.Protect, s.Middleware.RequireManager, s.CacheStats)

	// auth
	auth := v1.Group("/auth", s.Middleware.Protect)
	auth.Get("/me", s.Me)
	auth.Post("/logout", s.Logout)
	// restricted session may reach
	auth.Post("/oauth2/change-password", s.ChangePassword)

	// clients
	clients := v1.Group("/clients", s.Middleware.Protect, s.Middleware.RequireManager)
	clients.Get("/", s.Middleware.Authorization(model.MgrRightClientsAccess), s.ListClients)
	clients.Post("/", s.Middleware.Authorization(model.MgrRightClientsCreate), s.CreateClient)
	clients.Get("/:id", s.Middleware.Authorization(model.MgrRightClientsAccess), s.GetClient)
	clients.Patch("/:id", s.Middleware.Authorization(model.MgrRightClientsEdit), s.UpdateClient)
	clients.Delete("/:id", s.Middleware.Authorization(model.MgrRightClientsDelete), s.DeleteClient)

	// users
	users := v1.Group("/users", s.Middleware.Protect, s.Middleware.RequireManager)
	users.Get("/", s.Middleware.Authorization(model.MgrRightAccRead), s.ListUsers)
	users.Post("/", s.Middleware.Authorization(model.MgrRightAccManager), s.CreateUser)
	users.Get("/:login", s.Middleware.Authorization(model.MgrRightAccRead), s.GetUser)
	users.Patch("/:login", s.Middleware.Authorization(model.MgrRightAccManager), s.UpdateUser)
	users.Delete("/:login", s.Middleware.Authorization(model.MgrRightAccDelete), s.DeleteUser)

	// managers
	managers := v1.Group("/managers", s.Middleware.Protect, s.Middleware.RequireManager)
	managers.Get("/", s.Middleware.Authorization(model.MgrRightCfgManagers), s.ListManagers)
	managers.Post("/", s.Middleware.Authorization(model.MgrRightCfgManagers), s.CreateManager)
	managers.Get("/:login", s.Middleware.Authorization(model.MgrRightCfgManagers), s.GetManager)
	managers.Get("/:login/rights", s.Middleware.Authorization(model.MgrRightCfgManagers), s.GetManagerRights)
	managers.Patch("/:login", s.Middleware.Authorization(model.MgrRightCfgManagers), s.UpdateManager)
	managers.Delete("/:login", s.Middleware.Authorization(model.MgrRightCfgManagers), s.DeleteManager)
}
