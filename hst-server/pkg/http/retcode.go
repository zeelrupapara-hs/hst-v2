package http

import "github.com/gofiber/fiber/v2"

// RetCode is an MT5 return code. It travels in the response body so a client
// can tell "wrong password" from "must change password" without parsing text,
// while the HTTP status stays in the header.
//
// Values verified against the MT5 API reference, Return Codes, Authentication:
// support.metaquotes.net/en/docs/mt5/api/retcodes_authentication
type RetCode int

const (
	RetOK RetCode = 0

	// MT5 authentication block, 1000..1034. Only the codes this service can
	// actually produce are listed; the rest of the block covers certificates,
	// server identity and licensing, which are not our concerns.
	RetAuthClientInvalid   RetCode = 1000 // invalid type of the terminal
	RetAuthAccountInvalid  RetCode = 1001 // invalid account, our wrong password case
	RetAuthAccountDisabled RetCode = 1002 // the account is disabled
	RetAuthManagerNoConfig RetCode = 1011 // no manager configuration for the account
	RetAuthManagerIpBlock  RetCode = 1012 // ip address is not valid for the manager
	RetAuthServerBusy      RetCode = 1018 // the server is busy
	RetAuthAccountUnknown  RetCode = 1020 // unknown account
	RetAuthManagerType     RetCode = 1024 // this connection type is not permitted for manager
	RetAuthResetPassword   RetCode = 1026 // master password must be changed
	RetAuthOtpInvalid      RetCode = 1027 // invalid one time password
	RetAuthApiDisabled     RetCode = 1034 // api connection prohibited, USER_RIGHT_API_ENABLED

	// hst extensions. MT5 keeps a persistent authenticated connection and so
	// has no code for an expired bearer session or a locked out login. These
	// start at 60000 because the MT5 blocks run up to 15007.
	RetSessionExpired RetCode = 60001 // the session is gone, refresh or log in again
	RetAccountLocked  RetCode = 60002 // too many failed attempts, temporarily locked
)

// RetHTTPStatus maps a retcode onto the status a client should see.
// 1026 is the odd one: at login it accompanies a successful 200 that still
// hands over a token, while the middleware uses it to refuse. The refusing
// caller passes its own status, so this default is the login one.
func RetHTTPStatus(r RetCode) int {
	switch r {
	case RetOK, RetAuthResetPassword:
		return fiber.StatusOK
	case RetAuthClientInvalid, RetAuthAccountInvalid, RetAuthAccountUnknown,
		RetAuthOtpInvalid, RetSessionExpired:
		return fiber.StatusUnauthorized
	case RetAuthAccountDisabled, RetAuthManagerNoConfig, RetAuthManagerIpBlock,
		RetAuthManagerType, RetAuthApiDisabled, RetAccountLocked:
		return fiber.StatusForbidden
	case RetAuthServerBusy:
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
