package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SecurityHeaders sets the response headers that harden the admin panel.
func (m *Middleware) SecurityHeaders(c *fiber.Ctx) error {
	c.Set(fiber.HeaderXFrameOptions, "DENY")
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	c.Set(fiber.HeaderReferrerPolicy, "no-referrer")
	c.Set("Cross-Origin-Opener-Policy", "same-origin")

	// HSTS only makes sense on https, on http it would break local dev.
	if c.Protocol() == "https" {
		c.Set(fiber.HeaderStrictTransportSecurity, "max-age=31536000; includeSubDomains")
	}

	// swagger needs inline scripts and styles, the rest of the api serves json only.
	if !strings.HasPrefix(c.Path(), "/swagger") {
		c.Set(fiber.HeaderContentSecurityPolicy, "default-src 'none'; frame-ancestors 'none'")
	}

	return c.Next()
}

// CORS allows only the given origins, never a wildcard.
func (m *Middleware) CORS(origins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(origins, ","),
		AllowMethods:     strings.Join([]string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete, fiber.MethodOptions}, ","),
		AllowHeaders:     strings.Join([]string{fiber.HeaderAuthorization, fiber.HeaderContentType}, ","),
		AllowCredentials: true,
	})
}
