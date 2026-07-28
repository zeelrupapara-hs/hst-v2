package errors

import "errors"

const (
	BadRequest                   = "bad request"
	InternalServerError          = "internal server error"
	ServiceUnavailable           = "service unavailable"
	AlreadyExists                = "already exists"
	NotFound                     = "not found"
	Unauthorized                 = "unauthorized"
	Forbidden                    = "forbidden"
	Conflict                     = "conflict"
	BadQueryParams               = "invalid query params"
	TooManyRequests              = "too many requests"
	RequiredParams               = "param is required"
	MissingAuthorizationHeader   = "missing Authorization header"
	BasicAuth                    = "please provide valid basic credentials in the Authorization header"
	BearerToken                  = "please provide a valid bearer token in the Authorization header"
	InvalidToken                 = "expired or invalid token"
	InvalidSession               = "expired or invalid session"
	InvalidCredentials           = "incorrect login or password"
	AccountDisabled              = "the account is disabled"
	AccountLocked                = "the account is locked, try again later"
	NotAManager                  = "the login has no manager rights"
	TerminalNotPermitted         = "this terminal type is not permitted for the login"
	MustChangePassword           = "the password must be changed before continuing"
	UnauthorizedToAccessResource = "unauthorized to access this resource"
	CouldNotParseClientCfg       = "couldn't parse client stored config"
	EndpointNotFound             = "the endpoint you requested doesn't exist on server"
	DeleteWhileNotEmpty          = "you can't delete a record while it's not empty"
	SessionStoreUnavailable      = "the session store is unavailable"
	UnknownManagerRight          = "unknown manager right"
)

var (
	ErrBadRequest                   = errors.New(BadRequest)
	ErrInternalServerError          = errors.New(InternalServerError)
	ErrServiceUnavailable           = errors.New(ServiceUnavailable)
	ErrAlreadyExists                = errors.New(AlreadyExists)
	ErrNotFound                     = errors.New(NotFound)
	ErrUnauthorized                 = errors.New(Unauthorized)
	ErrForbidden                    = errors.New(Forbidden)
	ErrConflict                     = errors.New(Conflict)
	ErrTooManyRequests              = errors.New(TooManyRequests)
	ErrRequiredParams               = errors.New(RequiredParams)
	ErrMissingAuthorizationHeader   = errors.New(MissingAuthorizationHeader)
	ErrInvalidBasicAuth             = errors.New(BasicAuth)
	ErrInvalidBearerToken           = errors.New(BearerToken)
	ErrInvalidToken                 = errors.New(InvalidToken)
	ErrInvalidSession               = errors.New(InvalidSession)
	ErrInvalidCredentials           = errors.New(InvalidCredentials)
	ErrAccountDisabled              = errors.New(AccountDisabled)
	ErrAccountLocked                = errors.New(AccountLocked)
	ErrNotAManager                  = errors.New(NotAManager)
	ErrTerminalNotPermitted         = errors.New(TerminalNotPermitted)
	ErrMustChangePassword           = errors.New(MustChangePassword)
	ErrUnauthorizedToAccessResource = errors.New(UnauthorizedToAccessResource)
	ErrCouldNotParseClientCfg       = errors.New(CouldNotParseClientCfg)
	ErrEndpointNotFound             = errors.New(EndpointNotFound)
	ErrDeleteWhileNotEmpty          = errors.New(DeleteWhileNotEmpty)
	ErrSessionStoreUnavailable      = errors.New(SessionStoreUnavailable)
	ErrUnknownManagerRight          = errors.New(UnknownManagerRight)
)
