package middleware

import (
	"errors"
	"time"

	"hstserver/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// RequestsLogger will log every request with method, path, status and latency
func (m *Middleware) RequestsLogger(c *fiber.Ctx) error {
	start := time.Now()

	err := c.Next()

	// fiber sets the status in its error handler.
	status := c.Response().StatusCode()
	if err != nil {
		var fe *fiber.Error
		if errors.As(err, &fe) {
			status = fe.Code
		} else {
			status = fiber.StatusInternalServerError
		}
	}

	code := logger.CodeOK
	switch {
	case status >= 500:
		code = logger.CodeErr
	case status >= 400:
		code = logger.CodeWarn
	}

	fields := []any{
		"method", c.Method(),
		"path", c.Path(),
		"status", status,
		"latency_ms", time.Since(start).Milliseconds(),
		"ip", c.IP(),
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
	}

	m.Log.Log(logger.TypeNet, code, "request", fields...)

	return err
}
