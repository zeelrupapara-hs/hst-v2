package http

import (
	"crypto/rand"
	"math/big"
	"strconv"

	"hstserver/model"
	"hstserver/pkg/errors"
	"hstserver/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

// HttpResponse is the single envelope every endpoint returns.
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
// HttpResponseAccepted says the request was handed to the engine, not that it succeeded.
func (a *App) HttpResponseAccepted(c *fiber.Ctx, data interface{}) error {
	return c.Status(StatusAccepted).JSON(&HttpResponse{
		Success: true,
		Code:    RetOK,
		Data:    data,
	})
}

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
	setRetryAfter(c)
	return a.fail(c, StatusTooManyRequests, RetAuthServerBusy, ErrTooManyRequests, message)
}

// http 503
func (a *App) HttpResponseServiceUnavailable(c *fiber.Ctx, message error) error {
	setRetryAfter(c)
	return a.fail(c, StatusServiceUnavailable, RetAuthServerBusy, ErrServiceUnavailable, message)
}

// setRetryAfter tells the client to wait 3 to 7 whole seconds, jittered so rejected clients do not all come back together.
func setRetryAfter(c *fiber.Ctx) {
	wait := 3
	if n, err := rand.Int(rand.Reader, big.NewInt(5)); err == nil {
		wait += int(n.Int64())
	}
	c.Set(fiber.HeaderRetryAfter, strconv.Itoa(wait))
}

// http 500.
func (a *App) HttpResponseInternalServerErrorRequest(c *fiber.Ctx, message error) error {
	a.Log.Log(logger.TypeSys, logger.CodeErr, "request failed",
		"path", c.Path(), "method", c.Method(), "error", message.Error())

	return c.Status(StatusInternalServerError).JSON(&HttpResponse{
		Success: false,
		Code:    RetOK,
		Error:   ErrInternalServerError,
		Message: errors.InternalServerError,
	})
}

// HttpResponseStatus answers with the status a caller decided on rather than a fixed one.
func (a *App) HttpResponseStatus(c *fiber.Ctx, status int, message error) error {
	switch status {
	case StatusBadRequest:
		return a.HttpResponseBadRequest(c, message)
	case StatusForbidden:
		return a.HttpResponseForbidden(c, message)
	case StatusNotFound:
		return a.HttpResponseNotFound(c, message)
	case StatusConflict:
		return a.HttpResponseConflict(c, message)
	case StatusUnauthorized:
		return a.HttpResponseUnauthorized(c, message)
	case StatusTooManyRequests:
		return a.HttpResponseTooManyRequests(c, message)
	case StatusServiceUnavailable:
		return a.HttpResponseServiceUnavailable(c, message)
	default:
		return a.HttpResponseInternalServerErrorRequest(c, message)
	}
}

// HttpResponseRetCode answers 200 with a platform retcode.
func (a *App) HttpResponseRetCode(c *fiber.Ctx, code RetCode, data interface{}) error {
	return c.Status(StatusOK).JSON(&HttpResponse{
		Success: true,
		Code:    code,
		Data:    data,
	})
}

// HttpResponseDenied refuses a request while still carrying the platform retcode.
func (a *App) HttpResponseDenied(c *fiber.Ctx, status int, code RetCode, message error) error {
	errStr := ErrForbidden
	if status == StatusUnauthorized {
		errStr = ErrUnauthorized
	}

	// a refusal for load, not for the caller, so tell it when to come back
	if status == StatusServiceUnavailable || status == StatusTooManyRequests {
		setRetryAfter(c)
	}

	return a.fail(c, status, code, errStr, message)
}

// WS 200, the answer to an inbound frame.
func (a *App) WSResponseOK(event model.EventType, data any) *model.Event {
	return &model.Event{Type: event, Payload: encode(data)}
}

// WS 400
func (a *App) WSResponseBadRequest(event model.EventType, err error) *model.Event {
	return a.wsFail(model.EventBadRequest, event, err)
}

// WS 404
func (a *App) WSResponseNotFound(event model.EventType, err error) *model.Event {
	return a.wsFail(model.EventNotFound, event, err)
}

// WS 403
func (a *App) WSResponseForbidden(event model.EventType, err error) *model.Event {
	return a.wsFail(model.EventForbidden, event, err)
}

// WS 500
func (a *App) WSResponseInternalServerErrorRequest(event model.EventType, err error) *model.Event {
	return a.wsFail(model.EventInternalServerError, event, err)
}

// wsFail builds one refusal, naming the event that caused it.
func (a *App) wsFail(typ, event model.EventType, err error) *model.Event {
	message := ""
	if err != nil {
		message = err.Error()
		a.Log.Log(logger.TypeNet, logger.CodeWarn, "websocket request failed",
			"event", event, "error", message)
	}

	return &model.Event{
		Type:    typ,
		Payload: encode(model.ErrorPayload{Message: message, Reason: string(event)}),
	}
}

// encode is the event payload, which is carried raw.
func encode(data any) []byte {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return raw
}

// fail logs and writes one error body. message is a safe, caller-chosen string.
func (a *App) fail(c *fiber.Ctx, status int, code RetCode, errStr string, message error) error {
	res := &HttpResponse{Success: false, Code: code, Error: errStr}
	if message != nil {
		res.Message = message.Error()
	}
	return c.Status(status).JSON(res)
}
