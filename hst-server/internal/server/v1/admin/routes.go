package admin

import (
	"hstserver/model"

	"github.com/gofiber/fiber/v2"
)

// RegisterAdminV1 mounts the manager panel under /api/v1.
//
// Every route here is behind RequireManager and then a right, so a trading account cannot
// reach any of it however it asks.
func (s *Server) RegisterAdminV1(api fiber.Router) {
	v1 := api.Group("/v1")

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

	// floating leverage profiles
	leverages := v1.Group("/leverage-profiles", s.Middleware.Protect, s.Middleware.RequireManager)
	leverages.Get("/", s.Middleware.Authorization(model.MgrRightCfgGroups), s.ListLeverageProfiles)
	leverages.Post("/", s.Middleware.Authorization(model.MgrRightCfgGroups), s.CreateLeverageProfile)
	leverages.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.GetLeverageProfile)
	leverages.Put("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.UpdateLeverageProfile)
	leverages.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.DeleteLeverageProfile)

	// leverage rules and tiers
	leverages.Post("/:id/rules", s.Middleware.Authorization(model.MgrRightCfgGroups), s.CreateLeverageRule)
	leverages.Put("/:id/rules/reorder", s.Middleware.Authorization(model.MgrRightCfgGroups), s.ReorderLeverageRules)
	leverages.Put("/:id/rules/:ruleId", s.Middleware.Authorization(model.MgrRightCfgGroups), s.UpdateLeverageRule)
	leverages.Delete("/:id/rules/:ruleId", s.Middleware.Authorization(model.MgrRightCfgGroups), s.DeleteLeverageRule)

	// trading holidays
	holidays := v1.Group("/holidays", s.Middleware.Protect, s.Middleware.RequireManager)
	holidays.Get("/", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.ListHolidays)
	holidays.Post("/", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.CreateHoliday)
	holidays.Get("/check", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.CheckHoliday)
	holidays.Put("/reorder", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.ReorderHolidays)
	holidays.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.GetHoliday)
	holidays.Patch("/:id", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.UpdateHoliday)
	holidays.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgHolidays), s.DeleteHoliday)

	// universal symbols
	symbols := v1.Group("/symbols", s.Middleware.Protect, s.Middleware.RequireManager)
	symbols.Get("/", s.Middleware.Authorization(model.MgrRightCfgSymbols), s.ListSymbols)
	symbols.Post("/", s.Middleware.Authorization(model.MgrRightCfgSymbols), s.CreateSymbol)
	symbols.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgSymbols), s.GetSymbol)
	symbols.Patch("/:id", s.Middleware.Authorization(model.MgrRightCfgSymbols), s.UpdateSymbol)
	symbols.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgSymbols), s.DeleteSymbol)

	// groups
	groups := v1.Group("/groups", s.Middleware.Protect, s.Middleware.RequireManager)
	groups.Get("/", s.Middleware.Authorization(model.MgrRightCfgGroups), s.ListGroups)
	groups.Post("/", s.Middleware.Authorization(model.MgrRightCfgGroups), s.CreateGroup)
	groups.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.GetGroup)
	groups.Patch("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.UpdateGroup)
	groups.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgGroups), s.DeleteGroup)

	// group symbols
	groups.Get("/:id/symbols", s.Middleware.Authorization(model.MgrRightCfgGroups), s.ListGroupSymbols)
	groups.Post("/:id/symbols", s.Middleware.Authorization(model.MgrRightCfgGroups), s.CreateGroupSymbol)
	groups.Get("/:id/symbols/:symbolId", s.Middleware.Authorization(model.MgrRightCfgGroups), s.GetGroupSymbol)
	groups.Patch("/:id/symbols/:symbolId", s.Middleware.Authorization(model.MgrRightCfgGroups), s.UpdateGroupSymbol)
	groups.Delete("/:id/symbols/:symbolId", s.Middleware.Authorization(model.MgrRightCfgGroups), s.DeleteGroupSymbol)

	// group commissions
	groups.Get("/:id/commissions", s.Middleware.Authorization(model.MgrRightGroupCommission), s.ListGroupCommissions)
	groups.Post("/:id/commissions", s.Middleware.Authorization(model.MgrRightGroupCommission), s.CreateGroupCommission)
	groups.Get("/:id/commissions/:commissionId", s.Middleware.Authorization(model.MgrRightGroupCommission), s.GetGroupCommission)
	groups.Patch("/:id/commissions/:commissionId", s.Middleware.Authorization(model.MgrRightGroupCommission), s.UpdateGroupCommission)
	groups.Delete("/:id/commissions/:commissionId", s.Middleware.Authorization(model.MgrRightGroupCommission), s.DeleteGroupCommission)

	// server journal
	journal := v1.Group("/journal", s.Middleware.Protect, s.Middleware.RequireManager)
	journal.Get("/", s.Middleware.Authorization(model.MgrRightSrvJournals), s.MyJournal)
}
