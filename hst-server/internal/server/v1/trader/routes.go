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
	signin.Post("/forgot-password", s.ForgotPassword)
	signin.Post("/verify-code", s.VerifyCode)
	signin.Post("/reset-password", s.ResetPassword)

	// branding for the login screen, before a session exists
	root.Get("/public/broker", s.PublicBroker)

	trader := api.Group("/trader/v1", s.Middleware.Protect, s.Middleware.RequireTrader)
	trader.Get("/account", s.MyAccount)
	trader.Get("/profile", s.MyProfile)
	trader.Get("/symbols", s.MySymbols)
	// the literal paths must be registered before the :symbol wildcard below
	trader.Get("/symbols/tree", s.MySymbolsTree)
	trader.Get("/symbols/by_name", s.MySymbolByName)
	trader.Get("/symbols/:symbol/sessions", s.MySymbolSessions)
	trader.Get("/history", s.GetMyHistory)
	trader.Get("/history/positions", s.GetMyClosedPositions)
	trader.Get("/accounts/me/policies/ui", s.MyUIPolicies)
	trader.Get("/journal", s.GetMyJournal)

	// trading
	trader.Get("/orders", s.GetMyOrders)
	trader.Post("/orders", s.CreateMyOrder)
	trader.Put("/orders/:order_id", s.UpdateMyOrder)
	trader.Post("/orders/:order_id/cancel", s.CancelMyOrder)
	trader.Get("/positions", s.GetMyPositions)
	trader.Get("/positions/net", s.GetMyNetPositions)
	trader.Put("/positions/:position_id", s.UpdateMyPosition)
	trader.Post("/positions/:position_id/close", s.CloseMyPosition)
	trader.Post("/positions/:position_id/close-by", s.CloseByMyPosition)
	trader.Post("/requotes/:request_id/accept", s.AcceptMyRequote)
	trader.Get("/deals", s.GetMyDeals)

	// alerts
	trader.Get("/alerts", s.GetMyAlerts)
	trader.Post("/alerts", s.CreateMyAlert)
	trader.Get("/alerts/:alert_id", s.GetMyAlert)
	trader.Put("/alerts/:alert_id", s.UpdateMyAlert)
	trader.Delete("/alerts/:alert_id", s.DeleteMyAlert)

	// watchlists
	trader.Get("/watchlists", s.GetMyWatchlists)
	trader.Post("/watchlists", s.CreateMyWatchlist)
	trader.Put("/watchlists/:watchlist_id", s.UpdateMyWatchlist)
	trader.Delete("/watchlists/:watchlist_id", s.DeleteMyWatchlist)
	trader.Put("/watchlists/:watchlist_id/symbols", s.SetMyWatchlistSymbols)

	trader.Get("/mails", s.GetMyMails)
	trader.Get("/mails/:tracking_id", s.GetMyMail)
	trader.Post("/mails", s.SendMyMail)
	trader.Put("/mails/:tracking_id", s.UpdateMyDraft)
	trader.Delete("/mails/:tracking_id", s.DeleteMyMail)

	trader.Get("/news", s.GetMyNews)
	trader.Get("/news/:news_id", s.GetMyNewsItem)

	auth := trader.Group("/auth")
	auth.Get("/me", s.Me)
	auth.Post("/logout", s.Logout)
	auth.Post("/change-password", s.ChangePassword)
}
