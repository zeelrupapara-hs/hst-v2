package middleware

import (
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"

	"hstserver/pkg/http"
)

// Authorization gates a route on one manager right.
func (m *Middleware) Authorization(right uint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		snap, ok := utils.GetClient(c)
		if !ok {
			return m.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
		}

		if !snap.IsManager || !snap.ManagerRights.Has(right) {
			return m.App.HttpResponseForbidden(c, errs.ErrUnauthorizedToAccessResource)
		}

		return c.Next()
	}
}

// RequireTrader gates the trader panel: a session that owns one account and nothing else.
//
// Staff are refused here rather than let through with extra powers, because the two panels
// answer different questions and a manager reading a trader route would be reading it as
// somebody who has no account of their own.
func (m *Middleware) RequireTrader(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return m.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if snap.IsManager {
		return m.App.HttpResponseForbidden(c, errs.ErrNotATrader)
	}

	if !model.UsersRights(snap.Rights).CanConnect() {
		return m.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthAccountDisabled, errs.ErrAccountDisabled)
	}

	return c.Next()
}

// RequireManager gates a route on staff access alone, without a specific right.
func (m *Middleware) RequireManager(c *fiber.Ctx) error {
	snap, ok := utils.GetClient(c)
	if !ok {
		return m.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
	}

	if !snap.IsManager {
		return m.App.HttpResponseDenied(c, http.StatusForbidden, http.RetAuthManagerNoConfig, errs.ErrNotAManager)
	}

	return c.Next()
}
