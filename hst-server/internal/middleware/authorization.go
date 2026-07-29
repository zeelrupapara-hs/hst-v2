package middleware

import (
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
