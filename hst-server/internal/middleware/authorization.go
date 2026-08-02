package middleware

import (
	"hstserver/model"
	errs "hstserver/pkg/errors"
	"hstserver/utils"

	"github.com/gofiber/fiber/v2"

	"hstserver/pkg/http"
)

// Authorization gates a route: every right named must be held, because the platform makes some depend on others: reading an
// order needs account access as well as trade access.
func (m *Middleware) Authorization(rights ...uint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		snap, ok := utils.GetClient(c)
		if !ok {
			return m.App.HttpResponseInternalServerErrorRequest(c, errs.ErrCouldNotParseClientCfg)
		}

		if !snap.IsManager {
			return m.App.HttpResponseForbidden(c, errs.ErrUnauthorizedToAccessResource)
		}

		for _, right := range rights {
			if !snap.ManagerRights.Has(right) {
				return m.App.HttpResponseForbidden(c, errs.ErrUnauthorizedToAccessResource)
			}
		}

		return c.Next()
	}
}

// RequireTrader gates the trader panel.
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
