package middleware

import (
	"strings"

	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// HeaderReader pulls the bearer token and request metadata into Locals.
func (m *Middleware) HeaderReader(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		if token := strings.TrimSpace(authHeader[7:]); token != "" {
			c.Locals(http.LocalsToken, token)
		}
	}

	// c.IP() is the resolved client address.
	c.Locals(http.LocalsIp, c.IP())
	c.Locals(http.LocalsUserAgent, c.Get("User-Agent"))

	return c.Next()
}
