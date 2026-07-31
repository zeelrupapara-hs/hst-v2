package admin

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

// Login authenticates a member of staff and opens a manager session.
//
//	@Id				Login
//	@Description	Login with a manager account using basic auth
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"terminal type: 32 admin, 33 manager"
//	@Success		200		{object}	Response{data=ViewToken}
//	@Failure		401		{object}	Response
//	@Failure		403		{object}	Response
//	@Failure		500		{object}	Response
//	@Security		BasicAuth
//	@Router			/auth/v1/oauth2/login [post]
func (s *Server) Login(c *fiber.Ctx) error { return s.LoginAs(c, true) }

// RefreshToken exchanges a manager refresh token for a new pair.
//
//	@Id			RefreshToken
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		RefreshRequest	true	"the refresh token"
//	@Success	200		{object}	Response{data=ViewToken}
//	@Failure	401		{object}	Response
//	@Failure	500		{object}	Response
//	@Router		/auth/v1/oauth2/refresh [post]
func (s *Server) RefreshToken(c *fiber.Ctx) error { return s.RefreshSession(c, true) }

// Me describes the calling manager session.
//
//	@Id			Me
//	@Tags		Auth
//	@Produce	json
//	@Success	200	{object}	Response
//	@Failure	401	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/me [get]
func (s *Server) Me(c *fiber.Ctx) error { return s.CurrentSession(c) }

// Logout closes the calling manager session.
//
//	@Id			Logout
//	@Tags		Auth
//	@Produce	json
//	@Success	204	{object}	Response
//	@Failure	401	{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/logout [post]
func (s *Server) Logout(c *fiber.Ctx) error { return s.LogoutSession(c) }

// ChangePassword changes the calling manager's own password.
//
//	@Id			ChangePassword
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		ChangePasswordRequest	true	"the old and new passwords"
//	@Success	200		{object}	Response
//	@Failure	400		{object}	Response
//	@Failure	401		{object}	Response
//	@Security	BearerAuth
//	@Router		/api/v1/auth/oauth2/change-password [post]
func (s *Server) ChangePassword(c *fiber.Ctx) error { return s.SetPassword(c) }
