package errors

import "errors"

// The error vocabulary. One constant for the message, one var to compare
// against, so a handler never matches on a string literal.
const (
	BadRequest          = "bad request"
	InternalServerError = "internal server error"
	ServiceUnavailable  = "service unavailable"
	AlreadyExists       = "already exists"
	NotFound            = "not found"
	Unauthorized        = "unauthorized"
	Forbidden           = "forbidden"
	Conflict            = "conflict"
	BadQueryParams      = "invalid query params"
	TooManyRequests     = "too many requests"
	RequiredParams      = "param is required"
)

var (
	ErrBadRequest          = errors.New(BadRequest)
	ErrInternalServerError = errors.New(InternalServerError)
	ErrServiceUnavailable  = errors.New(ServiceUnavailable)
	ErrAlreadyExists       = errors.New(AlreadyExists)
	ErrNotFound            = errors.New(NotFound)
	ErrUnauthorized        = errors.New(Unauthorized)
	ErrForbidden           = errors.New(Forbidden)
	ErrConflict            = errors.New(Conflict)
	ErrBadQueryParams      = errors.New(BadQueryParams)
	ErrTooManyRequests     = errors.New(TooManyRequests)
	ErrRequiredParams      = errors.New(RequiredParams)
)
