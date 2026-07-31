package trader

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterTraderV1 mounts the trader panel under /api/trader/v1, plus the public signup.
//
// Nothing here is mask filtered: RequireTrader refuses a manager session, and every handler
// takes its login from the session, so there is nowhere to put somebody else's number.
func (s *Server) RegisterTraderV1(api fiber.Router, root fiber.Router) {
	// public signup, rate limited by ip like the login beside it
	root.Post("/auth/v1/register", s.Register)

	trader := api.Group("/trader/v1", s.Middleware.Protect, s.Middleware.RequireTrader)
	trader.Get("/account", s.MyAccount)
	trader.Get("/profile", s.MyProfile)
	trader.Get("/symbols", s.MySymbols)
}
