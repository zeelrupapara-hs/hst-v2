package middleware

import (
	"strings"

	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// HeaderReader pulls the bearer token and request metadata into Locals.
// It performs no I/O, so an unauthenticated route costs nothing.
func (m *Middleware) HeaderReader(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		if token := strings.TrimSpace(authHeader[7:]); token != "" {
			c.Locals(http.LocalsToken, token)
		}
	}

	c.Locals(http.LocalsIp, realIP(c))
	c.Locals(http.LocalsUserAgent, c.Get("User-Agent"))

	return c.Next()
}

// realIP prefers the proxy header, since fiber only trusts it when configured.
func realIP(c *fiber.Ctx) string {
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// the first entry is the originating client
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	return c.IP()
}
