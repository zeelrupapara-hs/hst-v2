package middleware

import (
	"strings"

	"hstserver/pkg/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// UpgradeWS rejects anything that is not a real upgrade, and finds the access token wherever the client was able to put it.
func (m *Middleware) UpgradeWS(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	// HeaderReader already took an Authorization header if there was one
	if _, ok := c.Locals(http.LocalsToken).(string); !ok {
		if token := tokenFromUpgrade(c); token != "" {
			c.Locals(http.LocalsToken, token)
		}
	}

	c.Locals("allowed", true)

	return c.Next()
}

// tokenFromUpgrade reads the token from the subprotocol, then the query.
func tokenFromUpgrade(c *fiber.Ctx) string {
	// "bearer, <token>": the client offers two subprotocols, the second being the credential
	if proto := c.Get("Sec-WebSocket-Protocol"); proto != "" {
		parts := strings.Split(proto, ",")
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "bearer") {
			if token := strings.TrimSpace(parts[1]); token != "" {
				return token
			}
		}
	}

	return strings.TrimSpace(c.Query("access_token"))
}
