package trader

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterTraderV1 mounts the trader panel under /api/trader/v1, plus the public signup.
func (s *Server) RegisterTraderV1(api, root fiber.Router) {
	// trading accounts sign in here, and a member of staff that tries is refused by the route it chose
	signin := root.Group("/auth/trader/v1")
	signin.Post("/login", s.Middleware.BasicAuthParser, s.Login)
	signin.Post("/refresh", s.RefreshToken)
	// public signup, rate limited by ip like the login beside it
	signin.Post("/register", s.Register)

	trader := api.Group("/trader/v1", s.Middleware.Protect, s.Middleware.RequireTrader)
	trader.Get("/account", s.MyAccount)
	trader.Get("/profile", s.MyProfile)
	trader.Get("/symbols", s.MySymbols)
	trader.Get("/history", s.GetMyHistory)

	// trading
	trader.Get("/orders", s.GetMyOrders)
	trader.Post("/orders", s.CreateMyOrder)
	trader.Put("/orders/:order_id", s.UpdateMyOrder)
	trader.Post("/orders/:order_id/cancel", s.CancelMyOrder)
	trader.Get("/positions", s.GetMyPositions)
	trader.Put("/positions/:position_id", s.UpdateMyPosition)
	trader.Post("/positions/:position_id/close", s.CloseMyPosition)
	trader.Post("/positions/:position_id/close-by", s.CloseByMyPosition)
	trader.Post("/requotes/:request_id/accept", s.AcceptMyRequote)
	trader.Get("/deals", s.GetMyDeals)

	auth := trader.Group("/auth")
	auth.Get("/me", s.Me)
	auth.Post("/logout", s.Logout)
	auth.Post("/change-password", s.ChangePassword)
}
