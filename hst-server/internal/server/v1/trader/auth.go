package trader

import (
	"github.com/gofiber/fiber/v2"

	v1 "hstserver/internal/server/v1"
)

// LoginRequest is the terminal asking, aliased for swagger.
type LoginRequest = v1.LoginRequest

// RefreshRequest is a refresh exchange, aliased for the same reason.
type RefreshRequest = v1.RefreshRequest

// ChangePasswordRequest is a password change, aliased for the same reason.
type ChangePasswordRequest = v1.ChangePasswordRequest

// ViewToken is an issued token pair, aliased for the same reason.
type ViewToken = v1.ViewToken

// Login authenticates a trading account and opens a trader session.
//
//	@Id				TraderLogin
//	@Description	Login with a trading account using basic auth. The investor password opens a read only session.
//	@Tags			Trader
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"terminal type: 0 client"
//	@Success		200		{object}	Response{data=ViewToken}
//	@Failure		401		{object}	Response
//	@Failure		403		{object}	Response
//	@Failure		500		{object}	Response
//	@Security		BasicAuth
//	@Router			/auth/trader/v1/login [post]
func (s *Server) Login(c *fiber.Ctx) error { return s.LoginAs(c, false) }

// RefreshToken exchanges a trader refresh token for a new pair.
//
//	@Id			TraderRefreshToken
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		RefreshRequest	true	"the refresh token"
//	@Success	200		{object}	Response{data=ViewToken}
//	@Failure	401		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/auth/trader/v1/refresh [post]
func (s *Server) RefreshToken(c *fiber.Ctx) error { return s.RefreshSession(c, false) }

// Me describes the calling trader session.
//
//	@Id			TraderMe
//	@Tags		Trader
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	401	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/auth/me [get]
func (s *Server) Me(c *fiber.Ctx) error { return s.CurrentSession(c) }

// Logout closes the calling trader session.
//
//	@Id			TraderLogout
//	@Tags		Trader
//	@Produce	json
//	@Success	204	{object}	Response
//	@Failure	401	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/auth/logout [post]
func (s *Server) Logout(c *fiber.Ctx) error { return s.LogoutSession(c) }

// ChangePassword changes the calling trader's own password.
//
//	@Id			TraderChangePassword
//	@Tags		Trader
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ChangePasswordRequest	true	"the old and new passwords"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	401		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/trader/v1/auth/change-password [post]
func (s *Server) ChangePassword(c *fiber.Ctx) error { return s.SetPassword(c) }
