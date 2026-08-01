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
	ClientHasFundedAccounts      = "this client has trading accounts holding funds, close or move them first"
	ManagerIpBlocked             = "this ip address is not permitted for the manager"
	LeverageNameExists           = "a leverage profile with this name already exists"
	LeverageLimitReached         = "the maximum of 1024 leverage profiles has been reached"
	LeverageRuleLimitReached     = "a leverage profile can hold at most 1024 rules"
	RuleNotInProfile             = "the rule does not belong to this leverage profile"
	ReorderMustListEveryRule     = "the reorder must list every rule of the profile exactly once"
	GroupAccessBeyondOwn         = "a manager cannot be granted group access beyond your own"
	RightsBeyondOwn              = "a manager cannot be granted a right you do not hold"
	GroupNotFound                = "no such group"
	RegistrationClosed           = "registration is not open for this account type"
	InvalidConnectionType        = "this connection type does not exist"
	EngineUnavailable            = "the trading engine did not answer"
	ReadOnlySession              = "this session may not trade"
	WrongPanel                   = "this login belongs to the other panel"
	AccountPending               = "the account is awaiting approval"
	NotATrader                   = "this panel is for trading accounts"
	HolidayDateInvalid           = "the month and day are not a real calendar date"
	ReorderMustListEveryHoliday  = "the reorder must list every holiday exactly once"
	RoutingAlreadyFirst          = "the routing rule is already first"
	RoutingAlreadyLast           = "the routing rule is already last"
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
	ErrClientHasFundedAccounts      = errors.New(ClientHasFundedAccounts)
	ErrManagerIpBlocked             = errors.New(ManagerIpBlocked)
	ErrLeverageNameExists           = errors.New(LeverageNameExists)
	ErrLeverageLimitReached         = errors.New(LeverageLimitReached)
	ErrLeverageRuleLimitReached     = errors.New(LeverageRuleLimitReached)
	ErrRuleNotInProfile             = errors.New(RuleNotInProfile)
	ErrReorderMustListEveryRule     = errors.New(ReorderMustListEveryRule)
	ErrGroupAccessBeyondOwn         = errors.New(GroupAccessBeyondOwn)
	ErrRightsBeyondOwn              = errors.New(RightsBeyondOwn)
	ErrGroupNotFound                = errors.New(GroupNotFound)
	ErrRegistrationClosed           = errors.New(RegistrationClosed)
	ErrInvalidConnectionType        = errors.New(InvalidConnectionType)
	ErrEngineUnavailable            = errors.New(EngineUnavailable)
	ErrReadOnlySession              = errors.New(ReadOnlySession)
	ErrWrongPanel                   = errors.New(WrongPanel)
	ErrAccountPending               = errors.New(AccountPending)
	ErrNotATrader                   = errors.New(NotATrader)
	ErrHolidayDateInvalid           = errors.New(HolidayDateInvalid)
	ErrReorderMustListEveryHoliday  = errors.New(ReorderMustListEveryHoliday)
	ErrRoutingAlreadyFirst          = errors.New(RoutingAlreadyFirst)
	ErrRoutingAlreadyLast           = errors.New(RoutingAlreadyLast)
)

// New is an error carrying a message the engine sent back, so a refusal reaches the client in
// the engine's own words.
func New(message string) error { return errors.New(message) }
