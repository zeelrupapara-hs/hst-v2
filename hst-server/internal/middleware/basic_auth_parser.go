package middleware

import (
	"encoding/base64"
	"strings"

	errs "hstserver/pkg/errors"
	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// BasicAuthParser decodes the basic auth header into Locals.
func (m *Middleware) BasicAuthParser(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return m.App.HttpResponseUnauthorized(c, errs.ErrMissingAuthorizationHeader)
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return m.App.HttpResponseUnauthorized(c, errs.ErrInvalidBasicAuth)
	}

	decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		return m.App.HttpResponseUnauthorized(c, errs.ErrInvalidBasicAuth)
	}

	username, password, found := strings.Cut(string(decoded), ":")
	if !found || username == "" || password == "" {
		return m.App.HttpResponseUnauthorized(c, errs.ErrInvalidBasicAuth)
	}

	c.Locals(http.LocalsUsername, username)
	c.Locals(http.LocalsPassword, password)

	return c.Next()
}
