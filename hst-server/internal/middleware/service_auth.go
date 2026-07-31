package middleware

import (
	"crypto/subtle"

	errs "hstserver/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

const serviceTokenHeader = "X-Service-Token"

// ServiceAuth validates internal worker requests using a shared service token.
func (m *Middleware) ServiceAuth(c *fiber.Ctx) error {
	token := m.Cfg.Internal.ServiceToken
	if token == "" {
		return m.App.HttpResponseForbidden(c, errs.ErrForbidden)
	}

	got := c.Get(serviceTokenHeader)
	if got == "" {
		return m.App.HttpResponseUnauthorized(c, errs.ErrUnauthorized)
	}

	if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
		return m.App.HttpResponseUnauthorized(c, errs.ErrUnauthorized)
	}
	return c.Next()
}
