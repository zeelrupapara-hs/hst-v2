package admin

import (
	"hstserver/model"

	"github.com/gofiber/fiber/v2"
)

// RegisterAdminV1 mounts the manager panel under /api/v1.
func (s *Server) RegisterAdminV1(api, root fiber.Router) {
	v1 := api.Group("/v1")

	// staff sign in here, and a trading account that tries is refused by the route it chose
	oauth := root.Group("/auth/v1")
	oauth.Post("/login", s.Middleware.BasicAuthParser, s.Login)
	oauth.Post("/refresh", s.RefreshToken)

	auth := v1.Group("/auth", s.Middleware.Protect, s.Middleware.RequireManager)
	auth.Get("/me", s.Me)
	auth.Post("/logout", s.Logout)
	// a restricted session may reach this one and nothing else
	auth.Post("/change-password", s.ChangePassword)

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

	// routing rules
	routing := v1.Group("/routing", s.Middleware.Protect, s.Middleware.RequireManager)
	routing.Get("/", s.Middleware.Authorization(model.MgrRightCfgRequests), s.ListRouting)
	routing.Post("/", s.Middleware.Authorization(model.MgrRightCfgRequests), s.CreateRouting)
	routing.Put("/order", s.Middleware.Authorization(model.MgrRightCfgRequests), s.ReorderRouting)
	routing.Post("/:id/move-up", s.Middleware.Authorization(model.MgrRightCfgRequests), s.MoveRoutingUp)
	routing.Post("/:id/move-down", s.Middleware.Authorization(model.MgrRightCfgRequests), s.MoveRoutingDown)
	routing.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgRequests), s.GetRouting)
	routing.Patch("/:id", s.Middleware.Authorization(model.MgrRightCfgRequests), s.UpdateRouting)
	routing.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgRequests), s.DeleteRouting)
	// the Dealers tab: who a "process to dealers" rule hands its requests to
	routing.Get("/:id/dealers", s.Middleware.Authorization(model.MgrRightCfgRequests), s.ListRoutingDealers)
	routing.Post("/:id/dealers", s.Middleware.Authorization(model.MgrRightCfgRequests), s.CreateRoutingDealer)
	routing.Delete("/:id/dealers/:login", s.Middleware.Authorization(model.MgrRightCfgRequests), s.DeleteRoutingDealer)
	routing.Put("/:id/dealers/:login/move", s.Middleware.Authorization(model.MgrRightCfgRequests), s.MoveRoutingDealer)

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

	// orders
	orders := v1.Group("/orders", s.Middleware.Protect, s.Middleware.RequireManager)
	orders.Get("/", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAllOrders)
	orders.Get("/accounts/:login", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAccountOrders)
	orders.Get("/:order_id", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetOrder)
	orders.Post("/", s.Middleware.Authorization(model.MgrRightTradesManager), s.CreateOrder)
	orders.Put("/:order_id", s.Middleware.Authorization(model.MgrRightTradesManager), s.UpdateOrder)
	orders.Post("/:order_id/cancel", s.Middleware.Authorization(model.MgrRightTradesManager), s.CancelOrder)

	// positions
	positions := v1.Group("/positions", s.Middleware.Protect, s.Middleware.RequireManager)
	positions.Get("/", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAllPositions)
	positions.Get("/accounts/:login", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAccountPositions)
	positions.Get("/:position_id", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetPosition)
	positions.Put("/:position_id", s.Middleware.Authorization(model.MgrRightTradesManager), s.UpdatePosition)
	positions.Post("/:position_id/close", s.Middleware.Authorization(model.MgrRightTradesManager), s.ClosePosition)
	positions.Post("/:position_id/close-by", s.Middleware.Authorization(model.MgrRightTradesManager), s.CloseByPosition)

	// deals
	deals := v1.Group("/deals", s.Middleware.Protect, s.Middleware.RequireManager)
	deals.Get("/", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAllDeals)
	deals.Get("/accounts/:login", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetAccountDeals)
	deals.Get("/:deal_id", s.Middleware.Authorization(model.MgrRightTradesRead), s.GetDeal)

	// the daily rollover: when it runs, and running it by hand
	eod := v1.Group("/system/end-of-day", s.Middleware.Protect, s.Middleware.RequireManager)
	eod.Get("/", s.Middleware.Authorization(model.MgrRightCfgTime), s.GetEndOfDay)
	eod.Put("/", s.Middleware.Authorization(model.MgrRightCfgTime), s.UpdateEndOfDay)
	eod.Post("/run", s.Middleware.Authorization(model.MgrRightCfgTime), s.RunEndOfDay)

	// chart history, for the manager panel
	v1.Get("/history", s.Middleware.Protect, s.Middleware.RequireManager,
		s.Middleware.Authorization(model.MgrRightSymbolDetails), s.GetHistory)

	// the dealing desk
	dealing := v1.Group("/dealing", s.Middleware.Protect, s.Middleware.RequireManager)
	dealing.Get("/", s.Middleware.Authorization(model.MgrRightTradesDealer), s.ListDealingRequests)
	// connecting as a dealer is what makes routing rules hand this manager work
	dealing.Get("/state", s.Middleware.Authorization(model.MgrRightTradesDealer), s.GetDealerState)
	dealing.Post("/connect", s.Middleware.Authorization(model.MgrRightTradesDealer), s.ConnectDealer)
	dealing.Post("/heartbeat", s.Middleware.Authorization(model.MgrRightTradesDealer), s.HeartbeatDealer)
	dealing.Post("/disconnect", s.Middleware.Authorization(model.MgrRightTradesDealer), s.DisconnectDealer)
	dealing.Post("/:request_id/confirm", s.Middleware.Authorization(model.MgrRightTradesDealer), s.ConfirmRequest)
	dealing.Post("/:request_id/requote", s.Middleware.Authorization(model.MgrRightTradesDealer), s.RequoteRequest)
	dealing.Post("/:request_id/reject", s.Middleware.Authorization(model.MgrRightTradesDealer), s.RejectRequest)
	dealing.Post("/:request_id/accept", s.Middleware.Authorization(model.MgrRightTradesDealer), s.AcceptRequote)
	dealing.Post("/:request_id/return", s.Middleware.Authorization(model.MgrRightTradesDealer), s.ReturnRequest)
	dealing.Post("/:request_id/cancel", s.Middleware.Authorization(model.MgrRightTradesDealer), s.CancelRequest)

	// balance operations, for the accountant only
	balance := v1.Group("/balance", s.Middleware.Protect, s.Middleware.RequireManager)
	balance.Post("/", s.Middleware.Authorization(model.MgrRightAccountant), s.CreateBalance)
	balance.Post("/deposit", s.Middleware.Authorization(model.MgrRightAccountant), s.CreateDeposit)
	balance.Post("/withdrawal", s.Middleware.Authorization(model.MgrRightAccountant), s.CreateWithdrawal)
	balance.Post("/credit", s.Middleware.Authorization(model.MgrRightAccountant), s.CreateCredit)
	balance.Post("/correction", s.Middleware.Authorization(model.MgrRightAccountant), s.CreateCorrection)

	// server journal
	journal := v1.Group("/journal", s.Middleware.Protect, s.Middleware.RequireManager)
	journal.Get("/", s.Middleware.Authorization(model.MgrRightSrvJournals), s.MyJournal)

	// datafeeds
	datafeeds := v1.Group("/datafeeds", s.Middleware.Protect, s.Middleware.RequireManager)
	datafeeds.Get("/", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ListDatafeeds)
	datafeeds.Get("/modules", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ListDatafeedModules)
	datafeeds.Post("/", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.CreateDatafeed)
	datafeeds.Get("/:id", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.GetDatafeed)
	datafeeds.Patch("/:id", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.UpdateDatafeed)
	datafeeds.Delete("/:id", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.DeleteDatafeed)
	datafeeds.Post("/:id/activate", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ActivateDatafeed)

	// datafeed symbols
	datafeeds.Get("/:id/symbols", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ListDatafeedSymbols)
	datafeeds.Post("/:id/symbols", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.CreateDatafeedSymbol)
	datafeeds.Get("/:id/symbols/resolve", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ResolveDatafeedSymbols)
	datafeeds.Get("/:id/symbols/:feedSymbolId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.GetDatafeedSymbol)
	datafeeds.Delete("/:id/symbols/:feedSymbolId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.DeleteDatafeedSymbol)

	// datafeed params
	datafeeds.Get("/:id/params", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ListDatafeedParams)
	datafeeds.Post("/:id/params", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.CreateDatafeedParam)
	datafeeds.Get("/:id/params/:paramId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.GetDatafeedParam)
	datafeeds.Patch("/:id/params/:paramId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.UpdateDatafeedParam)
	datafeeds.Delete("/:id/params/:paramId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.DeleteDatafeedParam)

	// datafeed translates
	datafeeds.Get("/:id/translates", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.ListDatafeedTranslates)
	datafeeds.Post("/:id/translates", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.CreateDatafeedTranslate)
	datafeeds.Get("/:id/translates/:translateId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.GetDatafeedTranslate)
	datafeeds.Patch("/:id/translates/:translateId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.UpdateDatafeedTranslate)
	datafeeds.Delete("/:id/translates/:translateId", s.Middleware.Authorization(model.MgrRightCfgDatafeeds), s.DeleteDatafeedTranslate)
}
