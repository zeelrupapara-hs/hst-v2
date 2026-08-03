package utils

import (
	"hstserver/pkg/cache"
	"hstserver/pkg/http"

	"github.com/gofiber/fiber/v2"
)

// GetClient returns the session snapshot Protect stored on the request.
func GetClient(c *fiber.Ctx) (*cache.Session, bool) {
	v, ok := c.Locals(http.LocalsClient).(*cache.Session)
	return v, ok
}

// GetToken returns the bearer token HeaderReader stored on the request.
func GetToken(c *fiber.Ctx) (string, bool) {
	v, ok := c.Locals(http.LocalsToken).(string)
	return v, ok && v != ""
}

// GetRealIP returns the client address HeaderReader resolved.
func GetRealIP(c *fiber.Ctx) string {
	if v, ok := c.Locals(http.LocalsIp).(string); ok {
		return v
	}
	return c.IP()
}

// GetUserAgent returns the raw user agent string.
func GetUserAgent(c *fiber.Ctx) string {
	if v, ok := c.Locals(http.LocalsUserAgent).(string); ok {
		return v
	}
	return c.Get("User-Agent")
}
