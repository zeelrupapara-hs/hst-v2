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

	// c.IP() is the resolved client address: the first valid entry of
	// X-Forwarded-For when the request came through a proxy listed in
	// TRUSTED_PROXIES, and the socket address otherwise. Reading the header
	// directly here instead would let any client that can reach the service
	// forge the address that lands in the session record and the journal.
	c.Locals(http.LocalsIp, c.IP())
	c.Locals(http.LocalsUserAgent, c.Get("User-Agent"))

	return c.Next()
}
