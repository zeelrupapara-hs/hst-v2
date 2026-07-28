package http

import (
	"errors"

	"hstserver/config"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

// Locals constants
const (
	LocalsAllowed = "allowed"
	LocalsClient  = "client"
	LocalsToken   = "token"
	LocalsDevice  = "device"
	LocalsOs      = "os"
	LocalsChannel = "channel"
)

const (
	StatusBadRequest          = fiber.StatusBadRequest
	StatusUnauthorized        = fiber.StatusUnauthorized
	StatusForbidden           = fiber.StatusForbidden
	StatusNotFound            = fiber.StatusNotFound
	StatusInternalServerError = fiber.StatusInternalServerError
	StatusOK                  = fiber.StatusOK
	StatusCreated             = fiber.StatusCreated
	StatusNoContent           = fiber.StatusNoContent
)

const (
	ErrBadRequest          = "Bad request"
	ErrInternalServerError = "Internal server error"
	ErrAlreadyExists       = "Already exists"
	ErrNotFound            = "Not Found"
	ErrUnauthorized        = "Unauthorized"
	ErrForbidden           = "Forbidden"
	ErrBadQueryParams      = "Invalid query params"
	ErrRequestTimeout      = "Request Timeout"
	ErrEndpointNotFound    = "The endpoint you requested doesn't exist on server"
)

type App struct {
	// fiber app instence
	*fiber.App
	// logger
	Log *logger.Logger
}

func NewApp(cfg *config.Config, log *logger.Logger) *App {
	newapp := fiber.New(fiber.Config{
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		EnableTrustedProxyCheck: true,
		AppName:                 "hstserver",
		// timeouts guard against slowloris
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		BodyLimit:    cfg.HTTP.BodyLimit,
		ErrorHandler: errorHandler(log),
	})

	return &App{
		App: newapp,
		Log: log,
	}
}

// errorHandler returns a generic body for 5xx, real detail only to the log.
func errorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		msg := ErrInternalServerError

		var fe *fiber.Error
		if errors.As(err, &fe) {
			code = fe.Code
			msg = fe.Message
		}

		if code >= fiber.StatusInternalServerError {
			log.Journal(logger.TypeSys, logger.CodeErr, "unhandled error",
				"path", c.Path(), "method", c.Method(), "error", err.Error())
			msg = ErrInternalServerError
		}

		return c.Status(code).JSON(fiber.Map{"error": msg})
	}
}
