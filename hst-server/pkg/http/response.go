package http

import (
	"hstserver/pkg/errors"
	"hstserver/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// HttpResponse is the single envelope every endpoint returns.
// Code is the MT5 retcode, not the HTTP status.
type HttpResponse struct {
	Success bool        `json:"success"`
	Code    RetCode     `json:"code"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
}

// http 200 ok
func (a *App) HttpResponseOK(c *fiber.Ctx, data interface{}) error {
	return c.Status(StatusOK).JSON(&HttpResponse{
		Success: true,
		Code:    RetOK,
		Data:    data,
	})
}

// http 201 created
func (a *App) HttpResponseCreated(c *fiber.Ctx, data interface{}) error {
	return c.Status(StatusCreated).JSON(&HttpResponse{
		Success: true,
		Code:    RetOK,
		Data:    data,
	})
}

// http 204 no content
func (a *App) HttpResponseNoContent(c *fiber.Ctx) error {
	return c.Status(StatusNoContent).JSON(&HttpResponse{
		Success: true,
		Code:    RetOK,
	})
}

// http 400 bad request
func (a *App) HttpResponseBadRequest(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusBadRequest, RetOK, ErrBadRequest, message)
}

// http 400 with the query string named as the cause
func (a *App) HttpResponseBadQueryParams(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusBadRequest, RetOK, ErrBadQueryParams, message)
}

// http 401
func (a *App) HttpResponseUnauthorized(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusUnauthorized, RetOK, ErrUnauthorized, message)
}

// http 403
func (a *App) HttpResponseForbidden(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusForbidden, RetOK, ErrForbidden, message)
}

// http 404
func (a *App) HttpResponseNotFound(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusNotFound, RetOK, ErrNotFound, message)
}

// http 409
func (a *App) HttpResponseConflict(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusConflict, RetOK, ErrConflict, message)
}

// http 429
func (a *App) HttpResponseTooManyRequests(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusTooManyRequests, RetAuthServerBusy, ErrTooManyRequests, message)
}

// http 503
func (a *App) HttpResponseServiceUnavailable(c *fiber.Ctx, message error) error {
	return a.fail(c, StatusServiceUnavailable, RetAuthServerBusy, ErrServiceUnavailable, message)
}

// http 500. The real error goes to the journal, never to the client, so a
// driver or query detail cannot leak through the body.
func (a *App) HttpResponseInternalServerErrorRequest(c *fiber.Ctx, message error) error {
	a.Log.Journal(logger.TypeSys, logger.CodeErr, "request failed",
		"path", c.Path(), "method", c.Method(), "error", message.Error())

	return c.Status(StatusInternalServerError).JSON(&HttpResponse{
		Success: false,
		Code:    RetOK,
		Error:   ErrInternalServerError,
		Message: errors.InternalServerError,
	})
}

// HttpResponseRetCode answers 200 with an MT5 retcode. A login that must change
// its password is still a success: it hands over a token and reports 1026.
func (a *App) HttpResponseRetCode(c *fiber.Ctx, code RetCode, data interface{}) error {
	return c.Status(StatusOK).JSON(&HttpResponse{
		Success: true,
		Code:    code,
		Data:    data,
	})
}

// HttpResponseDenied refuses a request while still carrying the MT5 retcode.
// The status is explicit because one code can mean two things: 1026 from login
// is a success, 1026 from the middleware is a refusal.
func (a *App) HttpResponseDenied(c *fiber.Ctx, status int, code RetCode, message error) error {
	errStr := ErrForbidden
	if status == StatusUnauthorized {
		errStr = ErrUnauthorized
	}
	return a.fail(c, status, code, errStr, message)
}

// fail logs and writes one error body. message is a safe, caller-chosen string.
func (a *App) fail(c *fiber.Ctx, status int, code RetCode, errStr string, message error) error {
	res := &HttpResponse{Success: false, Code: code, Error: errStr}
	if message != nil {
		res.Message = message.Error()
	}
	return c.Status(status).JSON(res)
}
